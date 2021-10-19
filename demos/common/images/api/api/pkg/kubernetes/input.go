package kubernetes

import (
	"encoding/json"
	log "logging"
)

type Command struct {
	Action    string
	Target    string
	Arguments []string
}

// TODO: Create a new way for argument checking

func ProcessInput(input []byte) Command {
	var commandTable Command
	err := json.Unmarshal(input, &commandTable)

	if err != nil {
		log.Println("Error on json parce:", err)
	}

	log.Println(commandTable)
	return commandTable
}
