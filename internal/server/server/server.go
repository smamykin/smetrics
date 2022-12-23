package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/storage"
	"log"
	"net/http"
)

func ListenAndServ() {
	r := NewRouter()
	log.Println("Starting the server.")
	err := http.ListenAndServe("127.0.0.1:8080", r)
	panic(err)
}

func NewRouter() chi.Router {
	r := chi.NewRouter()

	memStorage := storage.NewMemStorage()
	r.Method("POST", "/update/{metricType}/{metricName}/{metricValue}", &handlers.UpdateHandler{
		Repository: memStorage,
	})
	r.Method("GET", "/value/{metricType}/{metricName}", &handlers.GetHandler{
		Repository: memStorage,
	})
	r.Method("GET", "/", &handlers.ListHandler{
		Repository: memStorage,
	})

	return r
}
