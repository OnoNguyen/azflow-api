package main

import (
	"azflow-api/azure/auth"
	"azflow-api/azure/storage"
	"azflow-api/db"
	"azflow-api/domain/story"
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
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	// Set maximum request body size
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB
			next.ServeHTTP(w, r)
		})
	})

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

	// Add support for multipart forms
	srv.AddTransport(transport.MultipartForm{})

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

	// serve local ./video/[fileName] at http://localhost:8080/video/[fileName]
	r.HandleFunc("/video/{fileName}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		// fileName has the format [id].[ext], extract to id:int and ext:string
		fileName := vars["fileName"]
		parts := strings.Split(fileName, ".")
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "Invalid file name", http.StatusBadRequest)
			return
		}
		ext := parts[1]

		// find the file with the id not greater than and closest to the given id
		for i := id; i >= 0; i-- {
			if _, err := os.Stat(filepath.Join(RootDir, story.VideoWorkDir, fmt.Sprintf("%d.%s", i, ext))); err == nil {
				id = i
				break
			}
		}

		http.ServeFile(w, r, filepath.Join(RootDir, story.VideoWorkDir, fmt.Sprintf("%d.%s", id, ext)))
	}).Methods("GET")

	// handle file uploads
	r.HandleFunc("/video/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		// Set a maximum upload size
		const maxUploadSize = 10 << 20 // 10 MB
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		// Parse the multipart form
		err := r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			http.Error(w, "File too big or invalid request", http.StatusBadRequest)
			return
		}

		// Retrieve the file from the form
		file, header, err1 := r.FormFile("file")
		if err1 != nil {
			http.Error(w, "Unable to retrieve the file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// check to see if the file is an image,
		if !strings.Contains(header.Header.Get("Content-Type"), "image") {
			http.Error(w, "File is not an image", http.StatusBadRequest)
			return
		}
		// get ext
		ext := strings.Split(header.Filename, ".")[1]
		// only support png, jpg, jpeg
		if ext != "png" && ext != "jpg" && ext != "jpeg" {
			http.Error(w, "Image type not supported", http.StatusBadRequest)
			return
		}

		// Create the directory if it doesn't exist
		uploadDir := filepath.Join(RootDir, story.VideoWorkDir)

		// Create the destination file
		dst, err2 := os.Create(filepath.Join(uploadDir, fmt.Sprintf("%s.png", id)))
		if err2 != nil {
			http.Error(w, "Unable to save the file", http.StatusInternalServerError)
			fmt.Printf("Error: %v", err2)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the destination
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save the file", http.StatusInternalServerError)
			return
		}

		// Respond to the client
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("File uploaded successfully: %s", header.Filename)))
	}).Methods("POST")

	log.Printf("Connect to http://%s:%s/ for GraphQL playground", host, port)

	log.Fatal(http.ListenAndServe(host+":"+port, r))
}
