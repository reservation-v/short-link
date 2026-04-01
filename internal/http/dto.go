package http

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	URL      string `json:"url"`
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type resolveLinkResponse struct {
	URL string `json:"url"`
}

type errorResponse struct {
	Error string `json:"error"`
}
