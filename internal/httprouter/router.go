package httprouter

import (
	"github.com/go-chi/chi"
	"github.com/sgladkov/harvester/internal/interfaces"
)

var storage interfaces.Storage
var databaseDSN string

func MetricsRouter(s interfaces.Storage, dsn string) chi.Router {
	databaseDSN = dsn
	storage = s
	r := chi.NewRouter()
	r.Middlewares()
	r.Use(RequestLogger)
	r.Use(GzipHandle)
	r.Get("/", getAllMetrics)
	r.Get("/ping", ping)
	r.Post("/batchUpdate/", batchUpdate)
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
