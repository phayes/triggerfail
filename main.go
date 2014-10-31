package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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

	if args := flag.Args(); len(args) != 0 {
		root := args[0]
		rest := args[1:]
		cmd := exec.Command(root, rest...)
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
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
					log.Fatal(err)
				}
			}
			os.Stdout.Write(stdoutbuff[:n])

			// stderr
			n, err = stderrPipe.Read(stderrbuff)
			if err != nil {
				if err == io.EOF {
					done = true
				} else {
					log.Fatal(err)
				}
			}
			os.Stdout.Write(stderrbuff[:n])

			// If stdout contains a trigger, log it and possibly abort
			for _, Trigger := range Triggers {
				if strings.Contains(string(stdoutbuff), Trigger) {
					Failed = true
					if OptVerbose {
						log.Println("Command Failed. Found \"" + Trigger + "\"")
					}
					if OptAbort {
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
					Failed = true
					if OptVerbose {
						log.Println("Command Failed. Found \"" + Trigger + "\"")
					}
					if OptAbort {
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
			if err.Error() == "signal: killed" { // We killed the process intentionally, don't report it
				os.Exit(1)
			} else {
				log.Fatal(err)
			}
		}

		// Command complete, report pass / fail via exit status
		if Failed {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	} else {
		fmt.Println("checkfail let's you fail a program with an exit status of 1 if a string appears in it's output (either stderr or stdout). Use `checkfail --help` to see a list of available options.")
	}
}
