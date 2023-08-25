package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func analyzedump(targetPath string) string {
	//Example:  cdb -z .\071323-9906-01.dmp -c "!analyze -v; q" > dump.txt
	app := "cdb"

	arg0 := "-z"
	arg1 := targetPath
	arg2 := "-c"
	arg3 := "!analyze -v; q"

	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	leftpoint := strings.Index(string(stdout), "*******************************************************************************")
	rightpoint := strings.LastIndex(string(stdout), "quit:")

	// Return the output of the bug check analysis
	os.Remove(targetPath)
	return string(stdout[leftpoint:rightpoint])
}
