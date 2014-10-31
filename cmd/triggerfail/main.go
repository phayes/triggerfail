package main

import (
	"flag"
	"fmt"
	"github.com/phayes/triggerfail/triggerfail"
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
	cmd.Stdin = os.Stdin

	found, err := triggerfail.RunCommand(cmd, Triggers, os.Stdout, os.Stderr, OptAbort)
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
