package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// GetUserInput gets user input from stdin
func GetUserInput(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(message)
	text, _ := reader.ReadString('\n')
	text = strings.Trim(text, "\n")
	return text
}
