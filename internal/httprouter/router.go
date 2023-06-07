package httprouter

import (
	"database/sql"

	"github.com/go-chi/chi"
	"github.com/sgladkov/harvester/internal/interfaces"
)

var storage interfaces.Storage
var database *sql.DB
var key []byte

func MetricsRouter(s interfaces.Storage, db *sql.DB, k []byte) chi.Router {
	database = db
	storage = s
	key = k
	r := chi.NewRouter()
	r.Middlewares()
	//	r.Use(func(h http.Handler) http.Handler { return HandleHash(h, key) })
	r.Use(RequestLogger)
	r.Use(GzipHandle)
	r.Get("/", getAllMetrics)
	r.Get("/ping", ping)
	r.Post("/updates/", batchUpdate)
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
