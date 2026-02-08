package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/pawelpaszki/gorts-demo/internal/model"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

// BookHandler handles HTTP requests for books.
type BookHandler struct {
	service *service.BookService
}

// NewBookHandler creates a new book handler.
func NewBookHandler(svc *service.BookService) *BookHandler {
	return &BookHandler{service: svc}
}

// RegisterRoutes registers book routes on the given mux.
func (h *BookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/books", h.handleBooks)
	mux.HandleFunc("/api/books/", h.handleBook)
}

// handleBooks handles GET (list) and POST (create) for /api/books
func (h *BookHandler) handleBooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listBooks(w, r)
	case http.MethodPost:
		h.createBook(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleBook handles GET, PUT, DELETE for /api/books/{id}
func (h *BookHandler) handleBook(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/books/{id}
	id := strings.TrimPrefix(r.URL.Path, "/api/books/")
	if id == "" {
		http.Error(w, "Book ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getBook(w, r, id)
	case http.MethodPut:
		h.updateBook(w, r, id)
	case http.MethodDelete:
		h.deleteBook(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *BookHandler) listBooks(w http.ResponseWriter, r *http.Request) {
	books := h.service.ListBooks()
	respondJSON(w, http.StatusOK, books)
}

func (h *BookHandler) createBook(w http.ResponseWriter, r *http.Request) {
	var book model.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.service.CreateBook(&book); err != nil {
		if errors.Is(err, service.ErrInvalidBook) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrDuplicateISBN) {
			respondError(w, http.StatusConflict, "Book with this ISBN already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create book")
		return
	}

	respondJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) getBook(w http.ResponseWriter, r *http.Request, id string) {
	book, err := h.service.GetBook(id)
	if err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			respondError(w, http.StatusNotFound, "Book not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get book")
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func (h *BookHandler) updateBook(w http.ResponseWriter, r *http.Request, id string) {
	var book model.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	book.ID = id // Ensure ID matches path

	if err := h.service.UpdateBook(&book); err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			respondError(w, http.StatusNotFound, "Book not found")
			return
		}
		if errors.Is(err, service.ErrInvalidBook) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrDuplicateISBN) {
			respondError(w, http.StatusConflict, "Book with this ISBN already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update book")
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func (h *BookHandler) deleteBook(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.service.DeleteBook(id); err != nil {
		if errors.Is(err, service.ErrBookNotFound) {
			respondError(w, http.StatusNotFound, "Book not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete book")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// respondJSON writes a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
