package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
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

	// Create PR now
	prURL, err := createPR(baseRemoteName, baseRepoName, baseBranchName, compareRemoteName, compareBranchName)
	if err != nil {
		log.Fatal(err)
	}

	// Open created PR in the browser
	open.Run(prURL)
}

// createPR creates a PR and returns the URL to the PR
func createPR(
	owner string,
	repo string,
	base string,
	mergeRepo string,
	mergeBranch string,
) (string, error) {

	// get JIRA ticket number
	regex, _ := regexp.Compile(`(feature|bugfix)/([0-9]+)(-.*)?`)
	matched := regex.FindStringSubmatch(mergeBranch)
	if len(matched) == 0 {
		return "", fmt.Errorf("Not a valid branch to merge: %s", mergeBranch)
	}

	jiraID := matched[2]

	title, body := getTitleAndBody()

	// append issue number to the end of title
	prefix, _ := gitutil.ConfigString("story.core.issueprefix")
	title += fmt.Sprintf(" [%s%s]", prefix, jiraID)

	fmt.Println("Be patient...")

	head := fmt.Sprintf("%s:%s", mergeRepo, mergeBranch)

	requestURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	requestBody := strings.NewReader(fmt.Sprintf(`{ "title": "%s", "body": "%s", "head": "%s", "base": "%s" }`, title, body, head, base))

	req, err := http.NewRequest("POST", requestURL, requestBody)
	if err != nil {
		return "", err
	}

	token, _ := gitutil.ConfigString("story.core.oauthtoken")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.polaris-preview+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// get URL to the PR
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	prURL, err := extractPRURL(respBody)
	if err != nil {
		return "", err
	}

	return prURL, nil
}

func extractPRURL(respBody []byte) (string, error) {
	var f interface{}
	err := json.Unmarshal(respBody, &f)
	if err != nil {
		return "", err
	}
	m := f.(map[string]interface{})
	url := m["html_url"].(string)
	return url, nil
}

func getTitleAndBody() (string, string) {
	curDir, _ := os.Getwd()
	if _, err := os.Stat(fmt.Sprintf("%s/.git", curDir)); err != nil {
		log.Fatal(errors.New("Not a git repository"))
	}

	prTitleFile := fmt.Sprintf("%s/.git/PR_TITLE_MESSAGE", curDir)
	prBodyFile := fmt.Sprintf("%s/.git/PR_BODY_MESSAGE", curDir)

	fTitle, err := os.Create(prTitleFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fTitle.Close()

	fTitle.WriteString("Title")

	fBody, err := os.Create(prBodyFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fBody.Close()

	fBody.WriteString("Body")

	// Open the editor and let the user type in the title and body
	openEditor(prTitleFile)
	openEditor(prBodyFile)

	title, err := ioutil.ReadFile(prTitleFile)
	check(err)
	body, err := ioutil.ReadFile(prBodyFile)
	check(err)

	return string(title), string(body)
}

func openEditor(filename string) {
	cmd := exec.Command("vim", filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	check(err)
	err = cmd.Wait()
	check(err)
}
