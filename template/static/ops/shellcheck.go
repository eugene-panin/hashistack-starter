package main

import (
	"os/exec"
	"strings"
)

func checkScript(script string) (output string, installed bool, ok bool) {
	if _, err := exec.LookPath("shellcheck"); err != nil {
		return "", false, true
	}

	cmd := exec.Command("shellcheck", "-S", "warning", "-")
	cmd.Stdin = strings.NewReader(script)
	out, err := cmd.CombinedOutput()

	if err == nil {
		return "", true, true
	}
	if _, isExit := err.(*exec.ExitError); isExit {
		return string(out), true, false
	}
	return string(out) + "\nshellcheck did not run: " + err.Error(), true, false
}
