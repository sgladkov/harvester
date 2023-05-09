package httprouter

import (
	"github.com/go-chi/chi"
	"github.com/sgladkov/harvester/internal/interfaces"
)

var storage interfaces.Storage

func MetricsRouter(s interfaces.Storage) chi.Router {
	storage = s
	r := chi.NewRouter()
	r.Middlewares()
	r.Use(RequestLogger)
	r.Use(GzipHandle)
	r.Get("/", getAllMetrics)
	r.Route("/update/", func(r chi.Router) {
		r.Post("/", updateMetricJSON)
		r.Post("/{type}/{name}/{value}", updateMetric)
	})
	r.Route("/value/", func(r chi.Router) {
		r.Post("/", getMetricJSON)
		r.Get("/{type}/{name}", getMetric)
	})
	return r
}
