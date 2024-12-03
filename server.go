package main

import (
	"azflow-api/azure/auth"
	"azflow-api/azure/storage"
	"azflow-api/db"
	"azflow-api/gql"
	"azflow-api/openai"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

const (
	defaultPort = "8080"
	RootDir     = "./"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("API_HOST")
	port := os.Getenv("API_PORT")
	if port == "" {
		port = defaultPort
	}

	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true})

	r.Use(c.Handler, auth.Middleware())

	db.Init()
	db.Migrate()
	storage.Init()
	auth.Init()
	openai.Init()

	defer func() {
		err := db.CloseDB()
		if err != nil {
			log.Fatal(err)
		}
	}()

	srv := handler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: &gql.Resolver{}}))
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				log.Printf("CheckOrigin: %s", r.Host)
				// Check against your desired domains here
				return r.Host == "localhost:8080"
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})

	// Redirect handler
	r.HandleFunc("/s/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		fmt.Println("id is", id)
		fmt.Println("ShortURLs map contents:", gql.ShortURLs) // Debug print
		if link, ok := gql.ShortURLs[id]; ok {
			fmt.Println("Redirecting to:", link.LongURL) // Debug print
			http.Redirect(w, r, link.LongURL, http.StatusFound)
			return
		}
		fmt.Println("Short URL not found") // Debug print
		http.NotFound(w, r)
	})

	r.Handle("/", playground.Handler("GraphQL playground", "/gql"))
	r.Handle("/gql", srv)

	// serve local ./video/output.mp4 at http://localhost:8080/video/output.mp4
	r.HandleFunc("/video/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		log.Printf("Serving file: %s", r.URL.Path)
		http.ServeFile(w, r, RootDir+"/video/"+id)
	})

	log.Printf("Connect to http://%s:%s/ for GraphQL playground", host, port)

	log.Fatal(http.ListenAndServe(host+":"+port, r))
}
