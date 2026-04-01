package http

import stdhttp "net/http"

func NewRouter(handler *Handler) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	
	mux.HandleFunc("POST /links", handler.createLink)
	mux.HandleFunc("GET /links/{code}", handler.resolveLink)
	
	return mux
}
