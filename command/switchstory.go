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

	recent := c.Bool("recent")

	// Get repo instance
	root, _ := os.Getwd()
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
	}

	if recent {
		// if switching to most recent branch
		err = switchToMostRecentBranch(repo)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// otherwise, use pattern to find branches
		pattern := c.String("pattern")
		err = switchToBranch(repo, pattern)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func switchToBranch(repo *git.Repository, pattern string) error {

	branches, err := gitutil.FindBranches(repo, "^.*"+pattern+".*$", git.BranchLocal)
	if err != nil {
		return err
	}

	if len(branches) <= 0 {
		return fmt.Errorf("There are no branch to switch to")
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
		return err
	}

	// check if choosen number is within valid index
	if choice > len(brsNoHead) {
		return fmt.Errorf("Branch with choosen number: %d does not exist", choice)
	}

	branch := brsNoHead[choice-1]
	branchName, _ := branch.Name()

	err = doSwitch(repo, branchName)
	if err != nil {
		return err
	}

	return nil
}

func switchToMostRecentBranch(repo *git.Repository) error {

	branchName, err := gitutil.GetMostRecentBranch()
	if err != nil || branchName == "" {
		return fmt.Errorf("Could not find the most recent branch")
	}

	err = doSwitch(repo, branchName)
	if err != nil {
		return err
	}

	return nil
}

func doSwitch(repo *git.Repository, branchName string) error {

	fmt.Printf("\nSwitching to `%s`...\n", branchName)

	// Store current branch in most recent branch
	currentBranchName, err := gitutil.CurrentBranchName(repo)
	if err != nil {
		return err
	}
	err = gitutil.SetMostRecentBranch(currentBranchName)
	if err != nil {
		return err
	}

	// Stash all changes for current branch, if any
	fmt.Println("Stashing changes on current branch")
	err = gitutil.Stash(repo)
	if err != nil {
		return err
	}

	// checkout requested branch
	fmt.Printf("Checking out %s\n", branchName)
	err = gitutil.Checkout(repo, branchName)
	if err != nil {
		return err
	}

	// pop last stash if any
	fmt.Println("Popping last stashed changes for current branch")
	err = gitutil.PopLastStash(repo)
	if err != nil {
		return err
	}

	return nil
}
