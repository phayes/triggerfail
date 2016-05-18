package triggerfail

import (
	"fmt"
	"os/exec"
	"reflect"
	"testing"
)

func TestCommand(t *testing.T) {
	command := exec.Command("echo", "foo", "bar", "baz")
	triggers := []string{"foo"}
	triggered, err := RunCommand(command, triggers, Options{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(triggered, []string{"foo"}) {
		t.Error("Did not detect trigger")
	}

	command = exec.Command("echo", "foo", "bar", "baz")
	triggers = []string{"buzz"}
	triggered, err = RunCommand(command, triggers, Options{})
	if err != nil {
		t.Error(err)
	}
	if len(triggered) != 0 {
		t.Error("Detected trigger that doesn't exist")
		fmt.Println(triggered)
	}

	command = exec.Command("echo", "foo", "bar", "baz")
	triggers = []string{"foo", "baz"}
	triggered, err = RunCommand(command, triggers, Options{})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(triggered, []string{"foo", "baz"}) {
		t.Error("Did not detect trigger")
	}

	command = exec.Command("echo", "foo", "bar", "baz")
	triggers = []string{"foo", "baz"}
	triggered, err = RunCommand(command, triggers, Options{Abort: true})
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(triggered, []string{"foo"}) {
		t.Error("Did not abort early")
	}

	command = exec.Command("echo", "foo", "bar", "baz")
	triggers = []string{"foo"}
	triggered, err = RunCommand(command, triggers, Options{IgnoreStdOut: true})
	if err != nil {
		t.Error(err)
	}
	if len(triggered) != 0 {
		t.Error("Detected trigger that doesn't exist in stderr")
		fmt.Println(triggered)
	}

	command = exec.Command("command-that-doesnt-exist-asdf123")
	triggers = []string{"foo"}
	triggered, err = RunCommand(command, triggers, Options{})
	if err == nil {
		t.Error("Did not throw error when calling non-existent command")
	}
	if len(triggered) != 0 {
		t.Error("Detected trigger that doesn't exist")
		fmt.Println(triggered)
	}

}
