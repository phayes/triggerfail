package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	OptTrigger string
	OptAbort   bool
	OptVerbose bool
	Triggers   []string
	Failed     bool
)

func main() {
	flag.StringVar(&OptTrigger, "match", "", "Space seperate strings to match. If a match is found the exit code will be 0")
	flag.BoolVar(&OptAbort, "abort", false, "Abort a running command if a match is found. If abort is not passed the command is allowed to run to completion")
	flag.BoolVar(&OptVerbose, "v", false, "Verbose. Print the reason why we failed the command.")
	flag.Parse()

	// Parse the triggers
	Triggers = strings.Split(OptTrigger, " ")

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("triggerfail let's you fail a program with an exit status of 1 if a string appears in it's output (either stderr or stdout). Use `checkfail --help` to see a list of available options.")
		os.Exit(0)
	}

	root := args[0]
	rest := args[1:]
	cmd := exec.Command(root, rest...)

	found, err := RunCommand(cmd, Triggers, os.Stdout, os.Stderr, OptAbort)
	if OptVerbose && len(found) != 0 {
		for _, trig := range found {
			fmt.Println("Found trigger " + trig)
		}
	}
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			} else {
				// Unknown non-zero exit-status, exit with status 1
				os.Exit(1)
			}
		} else {
			// triggerfail failed internally for some reason
			log.Fatal(err.Error())
		}
	}
	if len(found) == 0 {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// Run a command, checking if triggers were found in it's output
// It returns a list of triggers found in the stdout and stderr of the running command
// This command takes the following arguments:
//   - *exec.Cmd - a pointer to an exec.Cmd (usually created by the exec.Command func). It can be configured as usual with the exception of stdin and stdout which should be left as is.
//   - []string  - A list of trigger strings. The command will be failed if it's stdout or stderr contains these strings.
//   - io.Writer - After consuming the stdout of the command to check for triggers, we pass along the results to this writer. Usually it would be set to os.Stdout or os.DevNull
//   - io.Writer - After consuming the stderr of the command to check for triggers, we pass along the results to this writer. Usually it would be set to os.Stderr or os.DevNull
//   - bool      - Abort the running commmand if a trigger word is found. If set to false the command will be allowed to run to completion, even with triggers.
func RunCommand(cmd *exec.Cmd, triggers []string, stdout io.Writer, stderr io.Writer, abort bool) ([]string, error) {
	found := make([]string, 0)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return found, err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return found, err
	}
	err = cmd.Start()
	if err != nil {
		return found, err
	}

	// Collect stdout and stderr
	stdoutbuff := make([]byte, 1024)
	stderrbuff := make([]byte, 1024)

	cmd.Run()

	for {
		//@@TODO: Move this to buffio.Scanner and two seperate co-running go-routines

		done := false

		// stdout
		n, err := stdoutPipe.Read(stdoutbuff)
		if err != nil {
			if err == io.EOF {
				done = true
			} else {
				return found, err
			}
		}
		stdout.Write(stdoutbuff[:n])

		// stderr
		n, err = stderrPipe.Read(stderrbuff)
		if err != nil {
			if err == io.EOF {
				done = true
			} else {
				return found, err
			}
		}
		stderr.Write(stderrbuff[:n])

		// If stdout contains a trigger, log it and possibly abort
		for _, Trigger := range Triggers {
			if strings.Contains(string(stdoutbuff), Trigger) {
				found = append(found, Trigger) //@@TODO: mutex
				if abort {
					if !done {
						cmd.Process.Kill() // abort the running command
					}
					done = true
				}
			}
		}

		// If stderr contains a trigger, log it and possibly abort
		for _, Trigger := range Triggers {
			if strings.Contains(string(stderrbuff), Trigger) {
				found = append(found, Trigger) //@@TODO: mutex
				if abort {
					if !done {
						cmd.Process.Kill() // abort the running command
					}
					done = true
				}
			}
		}

		if done {
			break
		}
	}

	err = cmd.Wait()
	if err != nil { // If we have an error abort everything
		if err.Error() == "signal: killed" { // We killed the process intentionally, don't report the error
			return found, nil
		} else {
			return found, err
		}
	}

	// Command complete
	return found, nil
}
