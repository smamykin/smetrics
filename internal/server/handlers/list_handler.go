package handlers

import (
	"html/template"
	"net/http"
)

type ListHandler struct {
	Repository IRepository
}

const listMetricsTmpl = `
<html>
    <ol>{{ range .GaugeMetrics }}
        <li>{{.Name}}:{{.Value}}</li>{{end}}
    </ol>
    <ol>{{ range .CounterMetrics }}
        <li>{{.Name}}:{{.Value}}</li>{{end}}
    </ol>
</html>`

func (l *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	w.Header().Set("Content-Type", "text/html")
	t, err := template.New("list").Parse(listMetricsTmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	gaugeMetrics, err := l.Repository.GetAllGauge()
	if err != nil {
		http.Error(w, "the error occurred while requesting the gauge metrics", http.StatusInternalServerError)
		return
	}

	counterMetrics, err := l.Repository.GetAllCounters()
	if err != nil {
		http.Error(w, "the error occurred while requesting the counter metrics", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, struct {
		GaugeMetrics   []GaugeMetric
		CounterMetrics []CounterMetric
	}{
		GaugeMetrics:   gaugeMetrics,
		CounterMetrics: counterMetrics,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
