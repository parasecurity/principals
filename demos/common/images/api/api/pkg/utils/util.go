package utils

import (
	"bufio"
	"fmt"
	"os"
)

func Prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func Int32Ptr(i int32) *int32 { return &i }
