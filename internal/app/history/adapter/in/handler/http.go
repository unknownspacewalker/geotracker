package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	"gitlab.com/spacewalker/locations/internal/app/history/core/port"
	"gitlab.com/spacewalker/locations/internal/pkg/errpack"
	"gitlab.com/spacewalker/locations/internal/pkg/util"
	"net/http"
	"time"
)

var (
	schemaDecoder = schema.NewDecoder()
)

// HTTPHandler is a handler that serves http requests.
type HTTPHandler struct {
	service port.HistoryService

	router *chi.Mux
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func NewHTTPHandler(service port.HistoryService) *HTTPHandler {
	router := chi.NewRouter()

	handler := &HTTPHandler{
		service: service,
		router:  router,
	}

	handler.setupRoutes()

	return handler
}

func (h *HTTPHandler) setupRoutes() {
	users := chi.NewRouter()

	users.Method(http.MethodGet, "/{username}/distance", http.HandlerFunc(h.getDistance))

	h.router.Mount("/users", users)
}

type getDistanceDTO struct {
	From string `schema:"from"`
	To   string `schema:"to"`
}

func (h *HTTPHandler) getDistance(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	err := r.ParseForm()
	if err != nil {
		status, body := errpack.ErrToHTTP(errpack.ErrInvalidArgument)
		util.Respond(w, status, body)
		return
	}

	var dto getDistanceDTO
	err = schemaDecoder.Decode(&dto, r.URL.Query())
	if err != nil {
		status, body := errpack.ErrToHTTP(errpack.ErrInvalidArgument)
		util.Respond(w, status, body)
		return
	}

	var fromPtr, toPtr *time.Time

	if dto.From != "" {
		from, err := time.Parse(time.RFC3339, dto.From)
		if err != nil {
			status, body := errpack.ErrToHTTP(errpack.ErrInvalidArgument)
			util.Respond(w, status, body)
			return
		}
		fromPtr = &from
	}

	if dto.From != "" {
		to, err := time.Parse(time.RFC3339, dto.To)
		if err != nil {
			status, body := errpack.ErrToHTTP(errpack.ErrInvalidArgument)
			util.Respond(w, status, body)
			return
		}
		toPtr = &to
	}

	var res port.HistoryServiceGetDistanceByUsernameResponse
	res, err = h.service.GetDistanceByUsername(r.Context(), port.HistoryServiceGetDistanceByUsernameRequest{
		Username: username,
		From:     fromPtr,
		To:       toPtr,
	})
	if err != nil {
		status, body := errpack.ErrToHTTP(err)
		util.Respond(w, status, body)
		return
	}

	util.Respond(w, http.StatusOK, res)
}
