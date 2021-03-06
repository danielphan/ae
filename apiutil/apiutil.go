package apiutil

import (
	"errors"
	"net/http"

	"github.com/verticalpalette/ae/logger"

	"appengine"
	"appengine/user"
)

var ErrMustLogIn = errors.New("Must be logged in.")

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	f(w, r)
}

func Error(f HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		c := appengine.NewContext(r)
		err := f(w, r)
		if err != nil {
			s := logger.Error(c, err)
			http.Error(w, s, http.StatusInternalServerError)
		}
		return err
	}
}

func Json(f HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Type", "application/json")
		return f(w, r)
	}
}

func Admin(f HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		c := appengine.NewContext(r)
		if !user.IsAdmin(c) {
			s := logger.Error(c, ErrMustLogIn)
			http.Error(w, s, http.StatusUnauthorized)
			return nil
		}
		return f(w, r)
	}
}
