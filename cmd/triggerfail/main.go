package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/phayes/triggerfail/triggerfail"
)

var (
	OptTrigger string
	OptAbort   bool
	OptVerbose bool
	OptStdErr  bool
	OptStdOut  bool
	Triggers   []string
	Failed     bool
)

func usage() {
	fmt.Println("triggerfail - fail a command with an exit status of 1 if a string appears in it's output (either stderr or stdout)")
	fmt.Println("")
	fmt.Println("USAGE")
	fmt.Println("  triggerfail \"<space-seperated-strings>\" [--abort] [-v] <command>")
	fmt.Println("")
	fmt.Println("OPTIONS")
	flag.PrintDefaults()
	fmt.Println("\nEXAMPLE")
	fmt.Println("  triggerfail --abort -v \"Error Warning\" mysqldump my_database > mysqlbackup.sql")
	os.Exit(0)
}

func main() {
	flag.BoolVar(&OptAbort, "abort", false, "Abort a running command if a match is found. If abort is not passed the command is allowed to run to completion")
	flag.BoolVar(&OptVerbose, "v", false, "Verbose. Print the reason why we failed the command.")
	flag.BoolVar(&OptStdErr, "stderr", false, "Only examine stderr for triggers.")
	flag.BoolVar(&OptStdOut, "stdout", false, "Only examine stdout for triggers.")

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) <= 1 {
		usage()
		os.Exit(0)
	}

	// Sanity check
	if OptStdOut && OptStdErr {
		fmt.Println("Cannot set both --stderr and --stdout at the same time.")
		os.Exit(1)
	}

	// Parse the triggers
	Triggers = strings.Split(args[0], " ")

	// Parse the command to run
	root := args[1]
	rest := args[2:]
	cmd := exec.Command(root, rest...)
	cmd.Stdin = os.Stdin

	opts := triggerfail.Options{
		Abort:        OptAbort,
		Stdout:       os.Stdout,
		Stderr:       os.Stderr,
		IgnoreStdErr: OptStdOut,
		IgnoreStdOut: OptStdErr,
	}

	found, err := triggerfail.RunCommand(cmd, Triggers, opts)
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
		os.Exit(82)
	}
}
