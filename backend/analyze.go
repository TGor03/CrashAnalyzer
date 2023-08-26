package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func analyzedump(targetPath string) string {
	//CDB must be installed in the !WINDOWS! system and added to the environment variable
	app := "cdb"
	arg0 := "-z"
	arg1 := targetPath
	arg2 := "-c"             //Execute command on launch
	arg3 := "!analyze -v; q" //Analyze file and then quit

	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	leftpoint := strings.Index(string(stdout), "*******************************************************************************")
	rightpoint := strings.LastIndex(string(stdout), "quit:")

	// Delete the dump file after analysis
	os.Remove(targetPath)

	// Return the output of the bug check analysis
	return string(stdout[leftpoint:rightpoint])
}
