package server

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"net/http"
	"time"
)

func NewRouter(repository handlers.IRepository, hashGenerator handlers.IHashGenerator, db *sql.DB) http.Handler {
	r := chi.NewRouter()

	r.Method("POST", "/update/{metricType}/{metricName}/{metricValue}", handlers.NewUpdateHandlerDefault(
		repository,
		ParameterBag{},
	))
	r.Method("GET", "/value/{metricType}/{metricName}", handlers.NewGetHandlerDefault(
		repository,
		ParameterBag{},
	))
	r.Method("GET", "/", &handlers.ListHandler{
		Repository: repository,
	})

	//region JSON-API
	r.Method("POST", "/update/", handlers.NewUpdateHandlerWithHashGenerator(repository, ParameterBag{}, hashGenerator, hashGenerator == nil))
	r.Method("POST", "/value/", handlers.NewGetHandlerWIthHashGenerator(repository, ParameterBag{}, hashGenerator, true))
	//endregion

	//region probe
	r.Method("GET", "/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	//endregion

	return gzipHandle(r)
}

type ParameterBag struct{}

func (p ParameterBag) GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
