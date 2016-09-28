package command

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/kidonchu/gitcli/gitutil"
)

// CmdNewStory creates new branch for story
func CmdNewStory(c *cli.Context) {

	branch := c.String("branch")
	if branch == "" {
		fmt.Println("Branch to create is not specified")
		return
	}

	from := c.String("source")
	source, err := gitutil.LookupBranchSource(from)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Get repo instance
	root, _ := os.Getwd()
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("* %s => %s\n", source, branch)

	answer := GetUserInput("Proceed with above items? (nY): ")
	if answer != "Y" {
		return
	}

	// Stash all changes for current branch, if any
	fmt.Println("Stashing changes on current branch")
	err = gitutil.Stash(repo)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Fetch from main repo before creating new branch
	fmt.Println("Fetching most recent branches")
	if err = gitutil.Fetch(repo, "ActiveCampaign"); err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Creating new branch")
	_, err = gitutil.CreateBranch(repo, branch, source)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Finding remote")
	remote, err := gitutil.GetRemote(repo, "origin")
	if err != nil {
		log.Fatalf("Unable to find remote: %+v\n", err)
		return
	}

	fmt.Println("Pushing to remote")
	ref := "refs/heads/" + branch
	err = gitutil.Push(repo, remote, ref)
	if err != nil {
		log.Fatal(err)
		return
	}
}
