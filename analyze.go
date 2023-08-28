package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func is_valid_dmp(targetPath string) bool {
	// Open the file
	file, err := os.Open(targetPath)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer file.Close() //Close the file when we're done

	// Read the first six bytes of the file
	var header [6]byte
	_, err = file.Read(header[:])
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	// Check if the first six bytes are equal to "PAGEDU"
	// MZ is the magic header for windows crashdump files
	if string(header[:]) == "PAGEDU" {
		return true
	}

	return false
}

func analyzedump(targetPath string) string {

	// Check if the file exists
	if _, err := os.Stat(targetPath); err != nil {
		fmt.Println(err.Error())
		return "Error while analyzing file"
	}

	//Check if the file is a PE file to prevent exploitation
	if !is_valid_dmp(targetPath) {
		return "Error while analyzing file (File is not a valid crashdump file)"
	}

	app := "cdb"
	arg0 := "-z"
	arg1 := targetPath
	arg2 := "-c"             //Execute command on launch
	arg3 := "!analyze -v; q" //Analyze file and then quit

	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return "Error while launching debugger agent (Is this a valid dump file?)"
	}

	leftpoint := strings.Index(string(stdout), "*******************************************************************************")
	rightpoint := strings.LastIndex(string(stdout), "quit:")

	// Delete the dump file after analysis
	os.Remove(targetPath)

	// Return the output of the bug check analysis
	return string(stdout[leftpoint:rightpoint])
}
