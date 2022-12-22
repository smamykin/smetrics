package handlers

import (
	"github.com/smamykin/smetrics/internal/server/storage"
	"html/template"
	"net/http"
)

type ListHandler struct {
	Repository IRepository
}

func (l *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "text/html")
	t, err := template.New("list").Parse("<html><ol>{{ range .GaugeMetrics }}<li>{{.Name}}:{{.Value}}</li>{{end}}</ol><ol>{{ range .CounterMetrics }}<li>{{.Name}}:{{.Value}}</li>{{end}}</ol></html>")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = t.Execute(w, struct {
		GaugeMetrics   []storage.GaugeMetric
		CounterMetrics []storage.CounterMetric
	}{
		GaugeMetrics:   l.Repository.GetAllGauge(),
		CounterMetrics: l.Repository.GetAllCounters(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
