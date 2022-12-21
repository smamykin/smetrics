package server

import (
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/storage"
	"net/http"
)

func ListenAndServ() {

	http.Handle("/update/", &handlers.UpdateHandler{
		Repository: storage.NewMemStorage(),
	})

	err := http.ListenAndServe("127.0.0.1:8080", nil)
	panic(err)
}
