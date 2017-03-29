package command

import (
	"encoding/json"
	"fmt"
)

type request struct {
	owner string
	repo  string
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

func (r *request) GetURL() (url string) {
	url = fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/pulls",
		r.owner, r.repo,
	)
	return
}

func (r *request) GetBody() (string, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	fmt.Printf("body = %+v\n", string(body))
	return string(body), nil
}
