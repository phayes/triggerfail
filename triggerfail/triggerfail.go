package triggerfail

import (
	"io"
	"os/exec"
	"strings"
)

// Run a command, checking if triggers were found in it's output.
// It returns a list of triggers found in the stdout and stderr of the running command.
// This command takes the following arguments:
//   cmd      *exec.Cmd - a pointer to an exec.Cmd (usually created by the exec.Command func). It can be configured as usual with the exception of stdin and stdout which should be left as is.
//   triggers []string  - A list of trigger strings. The command will be failed if it's stdout or stderr contains these strings.
//   stdout   io.Writer - After consuming the stdout of the command to check for triggers, we pass along the results to this writer. Usually it would be set to os.Stdout or os.DevNull
//   stderr   io.Writer - After consuming the stderr of the command to check for triggers, we pass along the results to this writer. Usually it would be set to os.Stderr or os.DevNull
//   abort    bool      - Abort the running commmand if a trigger word is found. If set to false the command will be allowed to run to completion, even with triggers.
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
		for _, trigger := range triggers {
			if strings.Contains(string(stdoutbuff), trigger) {
				found = append(found, trigger) //@@TODO: mutex
				if abort {
					if !done {
						cmd.Process.Kill() // abort the running command
					}
					done = true
				}
			}
		}

		// If stderr contains a trigger, log it and possibly abort
		for _, trigger := range triggers {
			if strings.Contains(string(stderrbuff), trigger) {
				found = append(found, trigger) //@@TODO: mutex
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
