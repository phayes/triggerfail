package triggerfail

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// Options provides configuration for RunCommand.
// An empty Options struct can be provided
type Options struct {
	Abort        bool      // Abort the running commmand if a trigger word is found. If set to false the command will be allowed to run to completion, even with triggers.
	Stdout       io.Writer // After consuming the stdout of the command to check for triggers, we pass along the results to this writer. Usually it would be set to os.Stdout. If not set stdout will be discarded.
	Stderr       io.Writer // After consuming the stderr of the command to check for triggers, we pass along the results to this writer. Usually it would be set to os.Stderr. If not set stderr will be discarded.
	IgnoreStdOut bool      // Don't evaluate StdOut for triggers.
	IgnoreStdErr bool      // Don't evaluate StdErr for triggers.
}

// RunCommand runs a command, checking if triggers were found in it's output.
// It returns a list of triggers found in the stdout and stderr of the running command.
// It takes the following arguments:
//   cmd      *exec.Cmd - a pointer to an exec.Cmd (usually created by the exec.Command func). It can be configured as usual with the exception of stdin and stdout which should be left as is.
//   triggers []string  - A list of trigger strings. The command will be failed if it's stdout or stderr contains these strings.
//   opts     Options   - Provide various config options for running this command. See Options struct.
func RunCommand(cmd *exec.Cmd, triggers []string, opts Options) ([]string, error) {
	var found []string

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return found, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return found, err
	}
	err = cmd.Start()
	if err != nil {
		return found, err
	}

	// When we are done, kill the process exactly once
	var done sync.Once

	// stdout
	go func() {
		stdoutScan := bufio.NewScanner(stdoutPipe)
		stdoutScan.Split(scanLines)

		for stdoutScan.Scan() {
			if opts.Stdout != nil {
				opts.Stdout.Write(stdoutScan.Bytes())
				opts.Stdout.Write([]byte("\n"))
			}
			if !opts.IgnoreStdOut {
				// If stdout contains a trigger, log it and possibly abort
				for _, trigger := range triggers {
					if strings.Contains(stdoutScan.Text(), trigger) {
						found = append(found, trigger)
						if opts.Abort {
							done.Do(func() { cmd.Process.Kill() }) // abort the running RunCommand
							return
						}
					}
				}
			}
		}
	}()

	// stderr
	go func() {
		stderrScan := bufio.NewScanner(stderrPipe)
		stderrScan.Split(scanLines)

		for stderrScan.Scan() {
			if opts.Stderr != nil {
				opts.Stderr.Write(stderrScan.Bytes())
				opts.Stderr.Write([]byte("\n"))
			}
			if !opts.IgnoreStdErr {
				// If stderr contains a trigger, log it and possibly abort
				for _, trigger := range triggers {
					if strings.Contains(stderrScan.Text(), trigger) {
						found = append(found, trigger)
						if opts.Abort {
							done.Do(func() { cmd.Process.Kill() }) // abort the running RunCommand
							return
						}
					}
				}
			}
		}
	}()

	err = cmd.Wait()
	if err != nil { // If we have an error abort everything
		if err.Error() == "signal: killed" { // We killed the process intentionally, don't report the error
			return found, nil
		}
		return found, err
	}

	// Command complete
	return found, nil
}

// scanLines is a split function for a Scanner that returns each line of
// text, but without stripping of any trailing end-of-line marker.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
