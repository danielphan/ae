package apiutil

import (
	"appengine"
	"appengine/datastore"
	"github.com/danielphan/ae/logger"
	"net/http"
)

type HandlerErrorFunc func(w http.ResponseWriter, r *http.Request) error

func Error(f HandlerErrorFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			s := logger.Error(appengine.NewContext(r), err)

			status := http.StatusInternalServerError
			if err == datastore.ErrNoSuchEntity {
				status = http.StatusNotFound
			}

			http.Error(w, s, status)
		}
	}
}

func Json(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f(w, r)
	}
}

func Wrap(f HandlerErrorFunc) http.HandlerFunc {
	return Json(Error(f))
}
