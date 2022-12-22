package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/smamykin/smetrics/internal/server/handlers"
	"github.com/smamykin/smetrics/internal/server/storage"
	"log"
	"net/http"
)

// V try to write a test on simple endpoint
// V write the tests on all the requirements
// V rewrite the handler of update on chi
// write new
// Сервер должен возвращать текущее значение запрашиваемой метрики в текстовом виде по запросу GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> (со статусом http.StatusOK).
// При попытке запроса неизвестной серверу метрики сервер должен возвращать http.StatusNotFound.
// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страничку со списком имён и значений всех известных ему на текущий момент метрик.
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
