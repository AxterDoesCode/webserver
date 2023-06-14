package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/AxterDoesCode/webserver/internal/database"
	"github.com/AxterDoesCode/webserver/pkg/apiconfig"
	"github.com/AxterDoesCode/webserver/pkg/middleware"
)

func main() {
	godotenv.Load()
	const port = "8080"
	r := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	jwtSecret := os.Getenv("JWT_SECRET")

	apiCfg := apiconfig.ApiConfig{
		FileserverHits: 0,
		JwtSecret:      jwtSecret,
	}

	db, err := database.NewDB(".")
	if err != nil {
		log.Fatal(err)
	}

	apiCfg.Database = *db

	r.Handle(
		"/app/*",
		apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))),
	)
	r.Handle(
		"/app",
		apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))),
	)
	r.Mount("/api", apiRouter)
	r.Mount("/admin", adminRouter)

	apiRouter.Get("/healthz", handlerReadiness)
	apiRouter.Post("/chirps", apiCfg.PostChirp)
	apiRouter.Get("/chirps", apiCfg.GetChirps)
	apiRouter.Get("/chirps/{chirpID}", apiCfg.GetChirpByID)
	apiRouter.Post("/users", apiCfg.AddUser)
	apiRouter.Post("/login", apiCfg.UserLogin)
	apiRouter.Put("/users", apiCfg.UpdateUserHandler)

	adminRouter.Get("/metrics", apiCfg.HandlerMetrics)
	corsr := middleware.MiddlewareCors(r)

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
