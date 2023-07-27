package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/akakream/MultiPlatform2IPFS/internal/ipfs"
	registry "github.com/akakream/MultiPlatform2IPFS/internal/registry"
)

type Server struct {
	baseURL       string
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

type CrdtPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

func NewServer(baseURL string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx

	return &Server{
		baseURL:       baseURL,
		quitch:        make(chan struct{}),
		cancelContext: cancel,
	}
}

func (s *Server) Start() {
    fmt.Println("Starting the MultiPlatform2IPFS server...")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// Publish a message to a topic
	r.Get("/health", makeHTTPHandler(s.handleHealth))
	r.Post("/image", makeHTTPHandler(s.handleCopy))

	go s.listenShutdown()

	go func() {
		if err := http.ListenAndServe(s.baseURL, r); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe Error: %v", err)
		}
	}()

    // Check if a local ipfs daeamon is running
    if !ipfs.DeamonIsUp() {
        fmt.Println("There is no local IPFS daemon is running! Uploads will fail!")
    }

    fmt.Println("MultiPlatform2IPFS server started.")

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
        // TODO: Gotta handle this properly on DistroMash
        cid = ""
    }

	resp := struct {
		Name string `json:"name"`
		Cid string `json:"cid"`
	}{
        Name: imageName,
        Cid: cid,
	}

	return writeJSON(w, http.StatusOK, resp)
}

/*
func postCid(imageName string, cid string) error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	distroMashURL, err := utils.GetEnv("DISTROMASH_URL", "localhost:3000")
	if err != nil {
        return err
	}

	crdtPair := CrdtPair{
		Key:   imageName,
		Value: cid,
	}
	payload, err := json.Marshal(crdtPair)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/api/v1/crdt", distroMashURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Non-OK HTTP status from the api with status code %d", resp.StatusCode)
	}
	return nil
}
*/

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
