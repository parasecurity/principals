package main
// This is part of cluster logging.
// The forward utility forwards stdin to
// the logging agent.
// TODO add flag for command tag

import (
	log "logging"
	"bufio"
	"os"
)

func main(){
	stdReader := bufio.NewReader(os.Stdin)
	for {
		msg, _ := stdReader.ReadString('\n')
		log.Println(msg)
	}
}
