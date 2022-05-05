package kubernetes

func Execute(command Command, registry *string) string {
	var result string
	if command.Action == "create" {
		Create(command, registry)
	} else if command.Action == "delete" {
		Delete(command)
	} else if command.Action == "execute" {
		Run(command, registry)
	} else if command.Action == "statistics" {
		Statistics(command)
	}
	result = "ok"
	return result
}
