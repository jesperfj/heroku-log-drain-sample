package main

import "net/http"

type handler func(w http.ResponseWriter, r *http.Request)

func checkAuth(correctPass string, pass handler) handler {

	return func(w http.ResponseWriter, r *http.Request) {

		_, password, ok := r.BasicAuth()

		if !ok {
			http.Error(w, "authtorization required", http.StatusBadRequest)
			return
		}

		if password != correctPass {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}
