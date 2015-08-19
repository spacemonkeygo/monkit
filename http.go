package monitor

import (
	"net/http"
)

func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p, contentType, found := PresentationFromPath(req.URL.Path)
	if !found {
		http.Error(w, "not found", 404)
		return
	}
	w.Header().Set("Content-Type", contentType)
	p(r, w)
}
