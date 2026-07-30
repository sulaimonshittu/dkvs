package main

import (
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	value, err := Get(key)
	if errors.Is(err, ErrorNonExistentKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	err = Put(key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.WritePut(key, string(value))
	w.WriteHeader(http.StatusCreated)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	err := Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.WriteDelete(key)
	w.WriteHeader(http.StatusOK)
}
