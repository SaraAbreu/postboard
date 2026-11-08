package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/SaraAbreu/postboard/api/internal/models"
	"github.com/SaraAbreu/postboard/api/internal/store"
)

type PostHandler struct {
	store  store.PostStore
	logger *slog.Logger
}

func NewPostHandler(s store.PostStore, l *slog.Logger) *PostHandler {
	return &PostHandler{store: s, logger: l}
}

// Routes registra las rutas usando el router nativo de Go (1.22+).
func (h *PostHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /posts", h.list)
	mux.HandleFunc("GET /posts/{id}", h.get)
	mux.HandleFunc("POST /posts", h.create)
	mux.HandleFunc("PUT /posts/{id}", h.update)
	mux.HandleFunc("DELETE /posts/{id}", h.delete)
}

func (h *PostHandler) list(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.List(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

func (h *PostHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	post, err := h.store.Get(r.Context(), id)
	if h.handleStoreErr(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func (h *PostHandler) create(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeAndValidate(w, r)
	if !ok {
		return
	}
	post, err := h.store.Create(r.Context(), in)
	if err != nil {
		h.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, post)
}

func (h *PostHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	in, ok := decodeAndValidate(w, r)
	if !ok {
		return
	}
	post, err := h.store.Update(r.Context(), id, in)
	if h.handleStoreErr(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func (h *PostHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if h.handleStoreErr(w, h.store.Delete(r.Context(), id)) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id inválido"})
		return 0, false
	}
	return id, true
}

func decodeAndValidate(w http.ResponseWriter, r *http.Request) (models.PostInput, bool) {
	var in models.PostInput
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return in, false
	}
	if errs := in.Validate(); len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return in, false
	}
	return in, true
}

func (h *PostHandler) handleStoreErr(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, store.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "post no encontrado"})
	} else {
		h.serverError(w, err)
	}
	return true
}

func (h *PostHandler) serverError(w http.ResponseWriter, err error) {
	h.logger.Error("error interno", "err", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
