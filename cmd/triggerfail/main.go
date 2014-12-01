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

func Usage() {
	fmt.Println("triggerfail - fail a command with an exit status of 1 if a string appears in it's output (either stderr or stdout)\n")
	fmt.Println("USAGE")
	fmt.Println("  triggerfail \"<space-seperated-strings>\" [--abort] [-v] <command>\n")
	fmt.Println("OPTIONS")
	flag.PrintDefaults()
	fmt.Println("\nEXAMPLE")
	fmt.Println("  triggerfail --abort -v \"Error Warning\" mysqldump my_database > mysqlbackup.sql")
	os.Exit(0)
}

func main() {
	flag.BoolVar(&OptAbort, "abort", false, "Abort a running command if a match is found. If abort is not passed the command is allowed to run to completion")
	flag.BoolVar(&OptVerbose, "v", false, "Verbose. Print the reason why we failed the command.")
	flag.Usage = Usage
	flag.Parse()

	args := flag.Args()

	if len(args) <= 1 {
		Usage()
		os.Exit(0)
	}

	// Parse the triggers
	Triggers = strings.Split(args[0], " ")

	// Parse the command to run
	root := args[1]
	rest := args[2:]
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
