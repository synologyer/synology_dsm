package types

type CommonResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID           string `json:"id"`
		RestartHTTPD bool   `json:"restart_httpd"`
	} `json:"data"`
}
