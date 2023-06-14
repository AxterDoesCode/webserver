package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/AxterDoesCode/webserver/internal/database"
)

func main() {
	godotenv.Load()
	const port = "8080"
	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	jwtSecret := os.Getenv("JWT_SECRET")

	apiCfg := apiConfig{
		fileserverHits: 0,
		jwtSecret:      jwtSecret,
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
	apiRouter.Post("/login", apiCfg.userLogin)
	apiRouter.Put("/users", apiCfg.updateUserHandler)

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
