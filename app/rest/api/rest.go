package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/theshamuel/gemini-proxy/app/rest"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Rest structure represents abstraction contains http server, exposed interface and version
type Rest struct {
	Service        restInterface
	Version        string
	httpServer     *http.Server
	DelayRequests  int
	TLSEnabled     bool
	CertPath       string
	PrivateKeyPath string
	lock           sync.Mutex
}

type restInterface interface {
	Send(request io.ReadCloser) ([]byte, error)
	GetMutex() *sync.Mutex
}

// Run http server
func (s *Rest) Run(port int) {
	log.Printf("[INFO] Run http server on port %d", port)
	s.lock.Lock()
	s.httpServer = s.buildHTTPServer(port, s.routes())
	s.lock.Unlock()
	var err error
	if s.TLSEnabled {
		err = s.httpServer.ListenAndServeTLS(s.CertPath, s.PrivateKeyPath)
	} else {
		err = s.httpServer.ListenAndServe()
	}

	if err != nil {
		log.Printf("[ERROR] Run http server on port %d failed: %v", port, err)
	}
	log.Printf("[WARN] http server terminated, %s", err)
}

// Shutdown http server
func (s *Rest) Shutdown() {
	log.Println("[WARN] shutdown http server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.lock.Lock()
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("[ERROR] http shutdown error, %s", err)
		}
		log.Println("[DEBUG] shutdown http server completed")
	}
	s.lock.Unlock()
}

func (s *Rest) buildHTTPServer(port int, router http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}

func (s *Rest) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Throttle(1000), middleware.RealIP, middleware.Recoverer, middleware.Logger)

	//corsMiddleware := cors.New(cors.Options{
	//	AllowedOrigins: []string{"*"},
	//	AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	//	AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	//	MaxAge:         300,
	//})

	corsMiddleware := cors.AllowAll()

	//health check api
	router.Use(corsMiddleware.Handler)
	router.Route("/", func(api chi.Router) {
		api.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(5, nil)))
		// nolint:revive
		api.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte(fmt.Sprintln("pong")))
			if err != nil {
				log.Printf("[ERROR] cannot write response: #%v", err)
			}
		})
	})

	router.Route("/api/", func(rapi chi.Router) {
		//app api
		rapi.Group(func(api chi.Router) {
			api.Use(middleware.Timeout(30 * time.Second))
			api.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(50, nil)))
			api.Use(middleware.NoCache)
			api.Post("/*", s.sendHandler)
		})
	})

	return router
}

// nolint:dupl
func (s *Rest) sendHandler(w http.ResponseWriter, r *http.Request) {

	if s.DelayRequests > 0 {
		s.Service.GetMutex().Lock()
		defer s.Service.GetMutex().Unlock()
		time.Sleep(1 * time.Second)
	}

	resp, err := s.Service.Send(r.Body)

	if err != nil {
		log.Printf("[ERROR] can not calculate min with error: %s", err.Error())
		rest.SendErrorJSON(w, r, http.StatusInternalServerError, err, rest.ErrServerInternal, "")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		log.Printf("[ERROR] can not write response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// DecodeJSON decodes a given reader into an interface using the json decoder.
func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(io.Discard, r) //nolint:errcheck
	return json.NewDecoder(r).Decode(v)
}
