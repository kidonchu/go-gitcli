package command

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/kidonchu/gitcli/gitutil"
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

// GetUserInputFromEditor opens an editor with given filename
// and when user finishes editing the file and exists, returns
// what user typed.
func GetUserInputFromEditor(msgFilename string) (string, error) {
	fi, err := os.Open(msgFilename)
	if err != nil {
		fi, err = os.Create(msgFilename)
	}
	if err != nil {
		return "", err
	}
	defer fi.Close()

	// Open the editor and let the user type
	editor, _ := gitutil.ConfigString("story.editor")
	cmd := exec.Command(editor, msgFilename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	// Read input from the message file
	msg, err := ioutil.ReadFile(msgFilename)
	if err != nil {
		return "", err
	}

	return strings.Trim(string(msg), " \n"), nil
}

func isGitRepo(dirPath string) bool {
	if _, err := os.Stat(fmt.Sprintf("%s/.git", dirPath)); err != nil {
		return false
	}
	return true
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
