package main

import (
	log "logging"
	"os/exec"
	"bytes"
)

func ntpSync(){
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd1 := exec.Command("ntpdate", "-4", "-b", "time.google.com")
	cmd1.Stdout = &out
	cmd1.Stderr = &stderr
	log.Println("Executing ", cmd1)
	err := cmd1.Run()
	if err != nil {
		log.Println(err, ": ", stderr.String())
		return
	}
	log.Println("Result: " + out.String())
}
