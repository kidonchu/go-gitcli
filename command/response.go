package command

type jsonError struct {
	Resource string `json:"resource"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type errorResponse struct {
	Message string      `json:"message"`
	Errors  []jsonError `json:"errors"`
}
