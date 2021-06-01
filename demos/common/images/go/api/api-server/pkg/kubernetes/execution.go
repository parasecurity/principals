package kubernetes

func Execute(command Command) string {
	var result string
	if command.Action == "create" {
		Create(command)
	} else if command.Action == "delete" {
		Delete(command)
	}
	result = "ok"
	return result
}
