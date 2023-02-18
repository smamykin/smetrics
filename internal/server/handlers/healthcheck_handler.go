package handlers

import "net/http"

func NewHealthcheckHandler(repositoryWithHealthCheck IRepositoryWithHealthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := repositoryWithHealthCheck.Healthcheck(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
