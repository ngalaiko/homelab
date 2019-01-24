package cloudflare

type listDomainsResponse struct {
	response
	Result []*domainResponse `json:"result"`
}

type response struct {
	Success bool             `json:"success"`
	Errors  []*errorResponse `json:"errors"`
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type domainResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}
