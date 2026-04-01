package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"

	"github.com/reservation-v/short-link/internal/domain"
	"github.com/reservation-v/short-link/internal/service"
)

const maxCreateLinkRequestBody = 4 << 10

type LinkService interface {
	Create(ctx context.Context, originalURL string) (service.CreateResult, error)
	Resolve(ctx context.Context, code string) (string, error)
}

type Handler struct {
	service LinkService
}

func NewHandler(service LinkService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) createLink(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req createLinkRequest
	body := stdhttp.MaxBytesReader(w, r.Body, maxCreateLinkRequestBody)
	defer body.Close()

	if err := decodeJSONBody(body, &req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidURL) {
			writeError(w, stdhttp.StatusBadRequest, "invalid request")
			return
		}

		writeError(w, stdhttp.StatusInternalServerError, "internal error")
		return
	}

	status := stdhttp.StatusOK
	if result.Created {
		status = stdhttp.StatusCreated
	}

	writeJSON(w, status, createLinkResponse{
		URL:      result.URL,
		Code:     result.Code,
		ShortURL: result.ShortURL,
	})
}

func (h *Handler) resolveLink(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	originalURL, err := h.service.Resolve(r.Context(), r.PathValue("code"))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCode):
			writeError(w, stdhttp.StatusBadRequest, "invalid code")
		case errors.Is(err, domain.ErrNotFound):
			writeError(w, stdhttp.StatusNotFound, "not found")
		default:
			writeError(w, stdhttp.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, stdhttp.StatusOK, resolveLinkResponse{URL: originalURL})
}

func decodeJSONBody(body io.Reader, out any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(out); err != nil {
		return err
	}

	// second JSON check
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("unexpected trailing json")
	}

	return nil
}
