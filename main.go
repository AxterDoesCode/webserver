package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/AxterDoesCode/webserver/internal/database"
)

func main() {
	const port = "8080"
	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	db, err := database.NewDB(".")
	if err != nil {
		log.Fatal(err)
	}

	apiCfg.database = *db

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
	apiRouter.Post("/chirps", apiCfg.postChirp)
	apiRouter.Get("/chirps", apiCfg.getChirps)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.getChirpByID)
	apiRouter.Post("/users", apiCfg.addUser)
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
