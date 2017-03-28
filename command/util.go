package command

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	git "github.com/libgit2/git2go"
)

var (
	branches []*git.Branch
	remotes  []*git.Remote
)

// GetUserInput gets user input from stdin
func GetUserInput(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(message)
	text, _ := reader.ReadString('\n')
	text = strings.Trim(text, "\n")
	return text
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
