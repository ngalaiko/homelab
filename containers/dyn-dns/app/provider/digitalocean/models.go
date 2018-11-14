package digitalocean

type response struct {
	Records []*record `json:"domain_records"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type record struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Data string `json:"data"`
}
