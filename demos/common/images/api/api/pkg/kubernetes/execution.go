package kubernetes

import (
	"net"
)

func Execute(command Command, registry *string, conn net.Conn) string {
	var result string
	if command.Action == "create" {
		Create(command, registry)
	} else if command.Action == "delete" {
		Delete(command)
	} else if command.Action == "execute" {
		Run(command, registry)
	} else if command.Action == "statistics" {
		Statistics(command, conn)
	}
	result = "ok"
	return result
}
