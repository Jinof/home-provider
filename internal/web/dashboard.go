package web

import (
	"net/http"
	"os"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", "no-store")

	file, err := os.ReadFile("./web/dist/index.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error loading dashboard"))
		return
	}

	w.Write(file)
}
