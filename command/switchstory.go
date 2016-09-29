package command

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/kidonchu/gitcli/gitutil"
	git "github.com/libgit2/git2go"
)

// CmdSwitchStory switches to another branch for story
func CmdSwitchStory(c *cli.Context) {

	pattern := c.String("pattern")

	// Get repo instance
	root, _ := os.Getwd()
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
		return
	}

	branches, err := gitutil.FindBranches(repo, "^.*"+pattern+".*$", git.BranchLocal)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(branches) <= 0 {
		fmt.Println("There are no branch to switch to. Exiting...")
		return
	}

	// filter out HEAD branch
	brsNoHead := branches[:0]
	for _, b := range branches {
		isHead, _ := b.IsHead()
		if !isHead { // don't display current head
			brsNoHead = append(brsNoHead, b)
		}
	}

	fmt.Printf("Choose branch to switch to:\n\n")
	for i, b := range brsNoHead {
		name, _ := b.Name()
		fmt.Printf("%d. %s\n", i+1, name)
	}

	answer := GetUserInput("\nBranch: ")
	choice, _ := strconv.Atoi(answer)
	if err != nil {
		log.Fatal(err)
		return
	}

	// check if choosen number is within valid index
	if choice > len(brsNoHead) {
		log.Fatalf("Branch with choosen number: %d does not exist", choice)
		return
	}

	branch := brsNoHead[choice-1]
	branchName, _ := branch.Name()

	fmt.Printf("\nSwitching to `%s`...\n", branchName)

	// Stash all changes for current branch, if any
	fmt.Println("Stashing changes on current branch")
	err = gitutil.Stash(repo)
	if err != nil {
		log.Fatal(err)
		return
	}

	// checkout requested branch
	fmt.Printf("Checking out %s\n", branchName)
	err = gitutil.Checkout(repo, branchName)
	if err != nil {
		log.Fatal(err)
		return
	}

	// pop last stash if any
	fmt.Println("Popping last stashed changes for current branch")
	err = gitutil.PopLastStash(repo)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Done")
}
