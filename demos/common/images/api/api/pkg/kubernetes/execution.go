package kubernetes

func Execute(command Command, registry *string) string {
	var result string
	if command.Action == "create" {
		Create(command, registry)
	} else if command.Action == "delete" {
		Delete(command)
	}
	result = "ok"
	return result
}
