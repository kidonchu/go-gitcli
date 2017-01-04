package command

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/kidonchu/gitcli/gitutil"
	"github.com/skratchdot/open-golang/open"
)

// CmdPullRequestStory switches to another branch for story
func CmdPullRequestStory(c *cli.Context) {

	from := c.String("source")
	source, err := gitutil.LookupBranchSource(from, true)
	if err != nil {
		log.Fatal(err)
	}

	// Extract source's remote and branch names
	var baseRemoteName, baseBranchName string
	sources := strings.Split(source, "/")
	if len(sources) == 1 {
		baseRemoteName = "origin"
		baseBranchName = sources[0]
	} else {
		baseRemoteName = sources[0]
		baseBranchName = sources[1]
	}

	baseRemoteURL, err := gitutil.ConfigString(fmt.Sprintf("remote.%s.url", baseRemoteName))
	if err != nil {
		log.Fatal(err)
	}

	// Extract base repo name
	regex, _ := regexp.Compile(`.*:(.*)/(.*)\.git?`)
	matched := regex.FindStringSubmatch(baseRemoteURL)
	baseRepoName := matched[2]

	// Get repo instance
	root, _ := os.Getwd()
	repo, err := gitutil.GetRepo(root)
	if err != nil {
		log.Fatal(err)
	}

	head, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}
	compareBranch := head.Branch()

	compareBranchName, err := compareBranch.Name()
	if err != nil {
		log.Fatal(err)
	}

	compareRemoteName, err := gitutil.ConfigString(fmt.Sprintf("branch.%s.remote", compareBranchName))
	if err != nil {
		log.Fatal(err)
	}

	compareRemoteURL, err := gitutil.ConfigString(fmt.Sprintf("remote.%s.url", compareRemoteName))
	if err != nil {
		log.Fatal(err)
	}

	// Extract compare remote name
	regex, _ = regexp.Compile(`.*:(.*)/(.*)\.git?`)
	matched = regex.FindStringSubmatch(compareRemoteURL)
	compareRemoteName = matched[1]

	url := fmt.Sprintf(
		"https://github.com/%s/%s/compare/%s...%s:%s?expand=1",
		baseRemoteName,
		baseRepoName,
		baseBranchName,
		compareRemoteName,
		compareBranchName,
	)

	open.Run(url)
}
