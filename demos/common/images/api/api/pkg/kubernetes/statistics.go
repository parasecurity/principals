package kubernetes

import (
	"log"

)

func Statistics(command Command) {
	if command.Target == "dns-stitching" {
		log.Printf("Reach here")
	}
	// TODO: Add a server connection to a specific port at dns-stitching pod
	// and pull specific statistics from the pod
}
