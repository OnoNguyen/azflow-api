package main

import (
	"azflow-api/graph"
	"azflow-api/handlers"
	database "azflow-api/internal/pkg/db/mysql"
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
	password := "password"
	handlers.HashPassord(password)

	router := chi.NewRouter()

	database.InitDB()
	defer func() {
		err := database.CloseDB()
		if err != nil {
			log.Fatal(err)
		}
	}()
	database.Migrate()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            true})

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router.Use(c.Handler)

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
