package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	const port = "8080"
	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	apiCfg := apiConfig{fileserverHits: 0}
	r.Handle(
		"/app/*",
		apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))),
	)
	r.Handle(
		"/app",
		apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))),
	)
	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Post("/validate_chirp", validateChirp)
	adminRouter.Get("/metrics", apiCfg.handlerMetrics)
	corsr := middlewareCors(r)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsr,
	}

	log.Printf("Serving on port : %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content=Type", "text-plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
