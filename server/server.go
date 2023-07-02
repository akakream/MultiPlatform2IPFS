package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	registry "github.com/akakream/MultiPlatform2IPFS/internal/registry"
)

type Server struct {
	port          string
	quitch        chan struct{}
	cancelContext context.CancelFunc
}

type apiError struct {
	Err    string `json:"err"`
	Status int    `json:"status"`
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type Image struct {
	Name string `json:"name"`
	Cid  string `json:"cid"`
}

func (e apiError) Error() string {
	return e.Err
}

func makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if e, ok := err.(apiError); ok {
				writeJSON(w, e.Status, e)
				return
			}
			writeJSON(w, http.StatusInternalServerError, apiError{Err: "internal server error", Status: http.StatusInternalServerError})
		}
	}
}

func NewServer(port string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx

	return &Server{
		port:          port,
		quitch:        make(chan struct{}),
		cancelContext: cancel,
	}
}

func (s *Server) Start() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// Publish a message to a topic
	r.Get("/health", makeHTTPHandler(s.handleHealth))
	r.Post("/image", makeHTTPHandler(s.handleCopy))

	go s.listenShutdown()

	go func() {
		if err := http.ListenAndServe(":"+s.port, r); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe Error: %v", err)
		}
	}()

	<-s.quitch
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apiError{Err: "invalid method", Status: http.StatusMethodNotAllowed}
	}
	return writeJSON(w, http.StatusOK, "OK")
}

func (s *Server) handleCopy(w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return apiError{Err: "invalid body", Status: http.StatusBadRequest}
	}
	defer r.Body.Close()

	var bodyJson Image
	if err := json.Unmarshal(body, &bodyJson); err != nil {
		log.Println(err)
		return apiError{Err: "body must be json", Status: http.StatusBadRequest}
	}

	imageName := bodyJson.Name
	if imageName == "" {
		return apiError{Err: "empty image name", Status: http.StatusBadRequest}
	}

	// Logic
	ctx := context.TODO()
	cid, err := registry.CopyImage(ctx, imageName)
	if err != nil {
		return apiError{Err: err.Error(), Status: http.StatusInternalServerError}
	}

	resp := Image{
		Name: imageName,
		Cid:  cid,
	}

	return writeJSON(w, http.StatusOK, resp)
}

func (s *Server) gracefullyQuitServer() {
	log.Println("Shutting down the server")

	// Shutdown services

	// Cancel the context
	s.cancelContext()
}

func (s *Server) listenShutdown() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	s.gracefullyQuitServer()
	close(s.quitch)
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
