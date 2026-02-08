package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

// AuthorHandler handles HTTP requests for authors.
type AuthorHandler struct {
	service *service.AuthorService
}

// NewAuthorHandler creates a new author handler.
func NewAuthorHandler(svc *service.AuthorService) *AuthorHandler {
	return &AuthorHandler{service: svc}
}

// RegisterRoutes registers author routes on the given mux.
func (h *AuthorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/authors", h.handleAuthors)
	mux.HandleFunc("/api/authors/", h.handleAuthor)
}

// handleAuthors handles GET (list) and POST (create) for /api/authors
func (h *AuthorHandler) handleAuthors(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAuthors(w, r)
	case http.MethodPost:
		h.createAuthor(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAuthor handles GET, PUT, DELETE for /api/authors/{id}
func (h *AuthorHandler) handleAuthor(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/authors/")
	if id == "" {
		http.Error(w, "Author ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getAuthor(w, r, id)
	case http.MethodPut:
		h.updateAuthor(w, r, id)
	case http.MethodDelete:
		h.deleteAuthor(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AuthorHandler) listAuthors(w http.ResponseWriter, r *http.Request) {
	// Check for country filter
	country := r.URL.Query().Get("country")
	var authors []*model.Author
	if country != "" {
		authors = h.service.GetAuthorsByCountry(country)
	} else {
		authors = h.service.ListAuthors()
	}
	respondJSON(w, http.StatusOK, authors)
}

func (h *AuthorHandler) createAuthor(w http.ResponseWriter, r *http.Request) {
	var author model.Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.service.CreateAuthor(&author); err != nil {
		if errors.Is(err, service.ErrInvalidAuthor) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create author")
		return
	}

	respondJSON(w, http.StatusCreated, author)
}

func (h *AuthorHandler) getAuthor(w http.ResponseWriter, r *http.Request, id string) {
	author, err := h.service.GetAuthor(id)
	if err != nil {
		if errors.Is(err, service.ErrAuthorNotFound) {
			respondError(w, http.StatusNotFound, "Author not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get author")
		return
	}

	respondJSON(w, http.StatusOK, author)
}

func (h *AuthorHandler) updateAuthor(w http.ResponseWriter, r *http.Request, id string) {
	var author model.Author
	if err := json.NewDecoder(r.Body).Decode(&author); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	author.ID = id

	if err := h.service.UpdateAuthor(&author); err != nil {
		if errors.Is(err, service.ErrAuthorNotFound) {
			respondError(w, http.StatusNotFound, "Author not found")
			return
		}
		if errors.Is(err, service.ErrInvalidAuthor) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update author")
		return
	}

	respondJSON(w, http.StatusOK, author)
}

func (h *AuthorHandler) deleteAuthor(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.DeleteAuthor(id); err != nil {
		if errors.Is(err, service.ErrAuthorNotFound) {
			respondError(w, http.StatusNotFound, "Author not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete author")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
