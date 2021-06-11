package kubernetes

import (
	"errors"
	"strings"
)

type Command struct {
	Action    string
	Name      string
	Arguments []string
}

func argCheck(input []string) error {
	// Create a new possible error
	err := errors.New("wrong input arguments")
	// Size of input array
	ArgLength := len(input)

	// Input arguments correctness checking
	if ArgLength < 2 {
		return err
	}

	if input[0] != "create" &&
		input[0] != "delete" {

		return err
	}

	if input[1] != "canary" &&
		input[1] != "canary-link" &&
		input[1] != "detector" &&
		input[1] != "detector-link" &&
		input[1] != "dga" &&
		input[1] != "analyser" &&
		input[1] != "snort" {

		return err
	}

	return nil
}

func lowerSplit(input string) []string {
	// Change input to lower case and split it
	inputLower := strings.ToLower(input)
	inputSplit := strings.Fields(inputLower)

	return inputSplit
}

func ProcessInput(input string) Command {
	commandTable := lowerSplit(input)
	err := argCheck(commandTable)

	if err != nil {
		err := Command{
			Action:    "Error",
			Name:      "",
			Arguments: nil,
		}
		return err
	}

	length := len(commandTable)
	newCommand := Command{
		Action:    commandTable[0],
		Name:      commandTable[1],
		Arguments: commandTable[2:length],
	}

	return newCommand
}
