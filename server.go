package main

import (
	"azflow-api/azure"
	"azflow-api/domain/auth"
	"azflow-api/graph"
	"github.com/joho/godotenv"

	//database "azflow-api/domain/pkg/db/mysql"
	database "azflow-api/domain/pkg/db/postgresql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            true})

	router.Use(c.Handler, auth.Middleware())

	database.Init()
	defer func() {
		err := database.CloseDB()
		if err != nil {
			log.Fatal(err)
		}
	}()
	database.Migrate()

	azure.Init()

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))
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

	router.Handle("/", playground.Handler("GraphQL playground", "/api"))
	router.Handle("/api", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)

	log.Fatal(http.ListenAndServe("localhost:"+port, router))
}
