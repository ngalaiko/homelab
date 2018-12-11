package api

import (
	"net/http"

	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
)

func (api *API) Routes() *mux.Router {
	// register routes
	r := mux.NewRouter()
	r.Handle("/collect", NewCollector(api.database)).Methods(http.MethodGet)

	r.Handle("/api/session", HandlerFunc(api.GetSession)).Methods(http.MethodGet)
	r.Handle("/api/session", HandlerFunc(api.CreateSession)).Methods(http.MethodPost)
	r.Handle("/api/session", HandlerFunc(api.DeleteSession)).Methods(http.MethodDelete)

	r.Handle("/api/sites", api.Authorize(HandlerFunc(api.GetSitesHandler))).Methods(http.MethodGet)
	r.Handle("/api/sites", api.Authorize(HandlerFunc(api.SaveSiteHandler))).Methods(http.MethodPost)
	r.Handle("/api/sites/{id:[0-9]+}", api.Authorize(HandlerFunc(api.SaveSiteHandler))).Methods(http.MethodPost)
	r.Handle("/api/sites/{id:[0-9]+}", api.Authorize(HandlerFunc(api.DeleteSiteHandler))).Methods(http.MethodDelete)

	r.Handle("/api/sites/{id:[0-9]+}/stats/site", api.Authorize(HandlerFunc(api.GetSiteStatsHandler))).Methods(http.MethodGet)
	r.Handle("/api/sites/{id:[0-9]+}/stats/site/agg", api.Authorize(HandlerFunc(api.GetAggregatedSiteStatsHandler))).Methods(http.MethodGet)
	r.Handle("/api/sites/{id:[0-9]+}/stats/site/realtime", api.Authorize(HandlerFunc(api.GetSiteStatsRealtimeHandler))).Methods(http.MethodGet)

	r.Handle("/api/sites/{id:[0-9]+}/stats/pages/agg", api.Authorize(HandlerFunc(api.GetAggregatedPageStatsHandler))).Methods(http.MethodGet)
	r.Handle("/api/sites/{id:[0-9]+}/stats/pages/agg/pageviews", api.Authorize(HandlerFunc(api.GetAggregatedPageStatsPageviewsHandler))).Methods(http.MethodGet)

	r.Handle("/api/sites/{id:[0-9]+}/stats/referrers/agg", api.Authorize(HandlerFunc(api.GetAggregatedReferrerStatsHandler))).Methods(http.MethodGet)
	r.Handle("/api/sites/{id:[0-9]+}/stats/referrers/agg/pageviews", api.Authorize(HandlerFunc(api.GetAggregatedReferrerStatsPageviewsHandler))).Methods(http.MethodGet)

	r.Handle("/health", HandlerFunc(api.Health)).Methods(http.MethodGet)

	// static assets & 404 handler
	box := packr.NewBox("./../../assets/build")
	r.Path("/tracker.js").Handler(serveTrackerFile(&box))
	r.Path("/").Handler(serveFileHandler(&box, "index.html"))
	r.Path("/index.html").Handler(serveFileHandler(&box, "index.html"))
	r.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(box)))
	r.NotFoundHandler = NotFoundHandler(&box)

	return r
}

func serveTrackerFile(box *packr.Box) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Tk", "N")
		next := serveFile(box, "js/tracker.js")
		next.ServeHTTP(w, r)
	})
}
