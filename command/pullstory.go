package command

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/kidonchu/gitcli/gitutil"
)

// CmdPullStory pulls specified source into current branch
func CmdPullStory(c *cli.Context) {

	from := c.String("source")
	source, err := gitutil.LookupBranchSource(from, true)
	if err != nil {
		log.Fatal(err)
	}

	// Get repo instance
	root, _ := os.Getwd()
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
	}

	var remoteName string

	// Extract source's remote and branch names
	sources := strings.Split(source, "/")
	if len(sources) == 1 {
		remoteName = "origin"
	} else {
		remoteName = sources[0]
	}

	// Fetch from repo before pulling
	fmt.Printf("Fetching most recent with remote: `%s`\n", remoteName)
	if err = gitutil.Fetch(repo, remoteName); err != nil {
		// do not fail entire app even if fetch fails
		log.Println(err)
	}

	fmt.Printf("Merging %s into local branch\n", source)
	err = gitutil.Pull(repo, source)
	if err != nil {
		log.Fatal(err)
	}
}
