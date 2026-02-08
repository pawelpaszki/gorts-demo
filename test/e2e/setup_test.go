package e2e

import (
	"net/http"
	"net/http/httptest"

	"github.com/pawelpaszki/gorts-demo/internal/handler"
	"github.com/pawelpaszki/gorts-demo/internal/middleware"
	"github.com/pawelpaszki/gorts-demo/internal/repository"
	"github.com/pawelpaszki/gorts-demo/internal/service"
)

// TestServer wraps all dependencies for e2e testing.
type TestServer struct {
	Server      *httptest.Server
	Mux         *http.ServeMux
	BookRepo    *repository.BookRepository
	AuthorRepo  *repository.AuthorRepository
	BookService *service.BookService
}

// NewTestServer creates a fully wired test server.
func NewTestServer() *TestServer {
	// Create repositories
	bookRepo := repository.NewBookRepository()
	authorRepo := repository.NewAuthorRepository()

	// Create services
	bookService := service.NewBookService(bookRepo)

	// Create handlers
	bookHandler := handler.NewBookHandler(bookService)
	healthHandler := handler.NewHealthHandler("1.0.0-test")

	// Setup routes
	mux := http.NewServeMux()
	bookHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)

	// Add root handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("Bookshelf API v1.0.0-test"))
	})

	// Wrap with middleware
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)

	// Create test server
	server := httptest.NewServer(h)

	return &TestServer{
		Server:      server,
		Mux:         mux,
		BookRepo:    bookRepo,
		AuthorRepo:  authorRepo,
		BookService: bookService,
	}
}

// Close shuts down the test server.
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// URL returns the base URL of the test server.
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// Client returns an HTTP client configured for the test server.
func (ts *TestServer) Client() *http.Client {
	return ts.Server.Client()
}
