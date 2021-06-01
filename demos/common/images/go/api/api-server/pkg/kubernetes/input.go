package kubernetes

import "strings"

type Command struct {
	Action string
	Name   string
}

// TODO: Check that action, type, name has a correct variable

func lowerSplit(input string) []string {
	// Change input to lower case and split it
	inputLower := strings.ToLower(input)
	inputSplit := strings.Fields(inputLower)

	return inputSplit
}

func ProcessInput(input string) Command {
	commandTable := lowerSplit(input)
	// When arguments are less than 3 return error
	// Example of right command:
	// 'create DeamonSet canary'
	if len(commandTable) != 2 {
		err := Command{
			Action: "Error",
			Name:   "",
		}
		return err
	}

	newCommand := Command{
		Action: commandTable[0],
		Name:   commandTable[1],
	}

	return newCommand
}
