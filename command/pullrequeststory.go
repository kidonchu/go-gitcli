package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
	realRemoteName := matched[1]
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
	prURL, err := createPR(realRemoteName, baseRepoName, baseBranchName, compareRemoteName, compareBranchName)
	if err != nil {
		fmt.Printf("err = %+v\n", err)
		return
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

	title, err := getTitle(mergeBranch)
	if err != nil {
		return "", err
	}

	body, err := getBody()
	if err != nil {
		return "", err
	}

	head := fmt.Sprintf("%s:%s", mergeRepo, mergeBranch)

	req := &request{
		owner: owner, repo: repo,
		Title: title, Body: body,
		Head: head, Base: base,
	}

	// Build request object
	httpreq, err := buildRequest(req)
	if err != nil {
		return "", err
	}

	// Send http request to create PR
	httpresp, err := http.DefaultClient.Do(httpreq)
	if err != nil {
		return "", err
	}
	defer httpresp.Body.Close()

	// get URL to the PR
	respBody, err := ioutil.ReadAll(httpresp.Body)
	if err != nil {
		return "", err
	}

	if httpresp.StatusCode != 200 {
		var resp errorResponse
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			return "", err
		}
		if resp.Message != "" {
			fmt.Printf("resp.Errors = %+v\n", resp.Errors)
			return "", errors.New(resp.Message)
		}
	}

	prURL, err := extractPRURL(respBody)
	if err != nil {
		return "", err
	}

	return prURL, nil
}

/**
 * getTitle asks the user to type in the title of the PR.
 * And then appends the issue number to the end of title
 * inside of brackets.
 */
func getTitle(branch string) (string, error) {
	curDir, _ := os.Getwd()
	if !isGitRepo(curDir) {
		return "", errors.New("Not a git repository")
	}

	msgFilename := fmt.Sprintf("%s/.git/PR_TITLE_MESSAGE", curDir)
	msg, err := GetUserInputFromEditor(msgFilename)
	if err != nil {
		return "", err
	}

	// get issue ticket number
	issueID, err := extractIssueNumber(branch)
	if err != nil {
		return "", err
	}

	// append issue number to the end of title
	prefix, _ := gitutil.ConfigString("story.issuePrefix")
	msg += fmt.Sprintf(" [%s%s]", prefix, issueID)

	return msg, nil
}

/**
* getBody asks the user to type in the body of the PR.
 */
func getBody() (string, error) {
	curDir, _ := os.Getwd()
	if !isGitRepo(curDir) {
		return "", errors.New("Not a git repository")
	}

	msgFilename := fmt.Sprintf("%s/.git/PR_BODY_MESSAGE", curDir)
	msg, err := GetUserInputFromEditor(msgFilename)
	if err != nil {
		return "", err
	}

	return msg, nil
}

/**
 * buildRequest creates a request object with
 * url, request body, and headers
 */
func buildRequest(req *request) (*http.Request, error) {

	url := req.GetURL()
	fmt.Printf("url = %+v\n", url)
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}

	httpreq, err := http.NewRequest(
		"POST", url, strings.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	token, _ := gitutil.ConfigString("story.oauthtoken")
	if token == "" {
		return nil, fmt.Errorf(
			"OAuth Token is required. Run '%s' to configure",
			"git config story.oauthtoken <oauth_token>",
		)

	}

	httpreq.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	httpreq.Header.Set("Accept", "application/vnd.github.polaris-preview+json")
	httpreq.Header.Set("Content-Type", "application/json")

	return httpreq, nil
}

/**
* extractIssueNumber extracts current issue's number
* using regex with given pattern. It returns the first matched
* string if matched, otherwise empty string.
 */
func extractIssueNumber(name string) (string, error) {

	pattern, _ := gitutil.ConfigString("story.issueBranchPattern")
	if pattern == "" {
		return "", fmt.Errorf(
			"Issue pattern required. Run '%s' to configure",
			"git config story.issueBranchPattern <issue_branch_pattern>",
		)
	}

	regex, _ := regexp.Compile(pattern)
	matched := regex.FindStringSubmatch(name)
	if len(matched) == 0 {
		return "", fmt.Errorf(
			"Unable to extract issue number from '%s' with the pattern '%s'",
			name, pattern,
		)
	}

	return matched[1], nil
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
