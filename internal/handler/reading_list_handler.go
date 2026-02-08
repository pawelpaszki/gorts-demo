package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

// ReadingListHandler handles HTTP requests for reading lists.
type ReadingListHandler struct {
	service *service.ReadingListService
}

// NewReadingListHandler creates a new reading list handler.
func NewReadingListHandler(svc *service.ReadingListService) *ReadingListHandler {
	return &ReadingListHandler{service: svc}
}

// RegisterRoutes registers reading list routes on the given mux.
func (h *ReadingListHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/lists", h.handleLists)
	mux.HandleFunc("/api/lists/", h.handleList)
}

// handleLists handles GET (list) and POST (create) for /api/lists
func (h *ReadingListHandler) handleLists(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listReadingLists(w, r)
	case http.MethodPost:
		h.createReadingList(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleList handles individual list operations: /api/lists/{id} and /api/lists/{id}/books/{bookId}
func (h *ReadingListHandler) handleList(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/lists/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "List ID required", http.StatusBadRequest)
		return
	}

	listID := parts[0]

	// Handle /api/lists/{id}/books/{bookId}
	if len(parts) >= 3 && parts[1] == "books" {
		bookID := parts[2]
		h.handleListBook(w, r, listID, bookID)
		return
	}

	// Handle /api/lists/{id}
	switch r.Method {
	case http.MethodGet:
		h.getReadingList(w, r, listID)
	case http.MethodPut:
		h.updateReadingList(w, r, listID)
	case http.MethodDelete:
		h.deleteReadingList(w, r, listID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListBook handles adding/removing books from a list
func (h *ReadingListHandler) handleListBook(w http.ResponseWriter, r *http.Request, listID, bookID string) {
	switch r.Method {
	case http.MethodPost:
		h.addBookToList(w, r, listID, bookID)
	case http.MethodDelete:
		h.removeBookFromList(w, r, listID, bookID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ReadingListHandler) listReadingLists(w http.ResponseWriter, r *http.Request) {
	lists := h.service.ListReadingLists()
	respondJSON(w, http.StatusOK, lists)
}

func (h *ReadingListHandler) createReadingList(w http.ResponseWriter, r *http.Request) {
	var list model.ReadingList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.service.CreateReadingList(&list); err != nil {
		if errors.Is(err, service.ErrInvalidReadingList) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create reading list")
		return
	}

	respondJSON(w, http.StatusCreated, list)
}

func (h *ReadingListHandler) getReadingList(w http.ResponseWriter, r *http.Request, id string) {
	list, err := h.service.GetReadingList(id)
	if err != nil {
		if errors.Is(err, service.ErrReadingListNotFound) {
			respondError(w, http.StatusNotFound, "Reading list not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get reading list")
		return
	}

	respondJSON(w, http.StatusOK, list)
}

func (h *ReadingListHandler) updateReadingList(w http.ResponseWriter, r *http.Request, id string) {
	var list model.ReadingList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	list.ID = id

	if err := h.service.UpdateReadingList(&list); err != nil {
		if errors.Is(err, service.ErrReadingListNotFound) {
			respondError(w, http.StatusNotFound, "Reading list not found")
			return
		}
		if errors.Is(err, service.ErrInvalidReadingList) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update reading list")
		return
	}

	respondJSON(w, http.StatusOK, list)
}

func (h *ReadingListHandler) deleteReadingList(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.DeleteReadingList(id); err != nil {
		if errors.Is(err, service.ErrReadingListNotFound) {
			respondError(w, http.StatusNotFound, "Reading list not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete reading list")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ReadingListHandler) addBookToList(w http.ResponseWriter, r *http.Request, listID, bookID string) {
	if err := h.service.AddBookToList(listID, bookID); err != nil {
		if errors.Is(err, service.ErrReadingListNotFound) {
			respondError(w, http.StatusNotFound, "Reading list not found")
			return
		}
		if errors.Is(err, service.ErrBookNotFound) {
			respondError(w, http.StatusNotFound, "Book not found")
			return
		}
		if errors.Is(err, service.ErrBookAlreadyInList) {
			respondError(w, http.StatusConflict, "Book already in list")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to add book to list")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ReadingListHandler) removeBookFromList(w http.ResponseWriter, r *http.Request, listID, bookID string) {
	if err := h.service.RemoveBookFromList(listID, bookID); err != nil {
		if errors.Is(err, service.ErrReadingListNotFound) {
			respondError(w, http.StatusNotFound, "Reading list not found")
			return
		}
		if errors.Is(err, service.ErrBookNotInList) {
			respondError(w, http.StatusNotFound, "Book not in list")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to remove book from list")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
