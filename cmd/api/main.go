package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type RequestData struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	N_Reports int    `json:"n_reports"`
}

func exportPngHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var requests []RequestData
	err := json.NewDecoder(r.Body).Decode(&requests)
	if err != nil {
		fmt.Println("Error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	values := make(plotter.Values, 0, len(requests))
	for _, value := range requests {
		values = append(values, float64(value.N_Reports))
	}
	p := plot.New()
	p.Title.Text = "Test"
	p.Y.Label.Text = "Número de reportes"
	bar, err := plotter.NewBarChart(values, vg.Points(20))
	if err != nil {
		log.Fatal(err)
	}
	p.Add(bar)
	if err := p.Save(5*vg.Inch, 3*vg.Inch, "barchart.pdf"); err != nil {
		panic(err)
	}
}

func corsHandler(h http.Handler) http.HandlerFunc {
	allowed := map[string]bool{
		"http://localhost:3000": true,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Solo permitir si viene desde un origen permitido
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		// Preflight
		if r.Method == http.MethodOptions {
			// Si el origin no está permitido, puedes responder 403
			if origin != "" && !allowed[origin] {
				http.Error(w, "CORS not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Si viene con Origin (browser) y no está permitido, bloquea
		if origin != "" && !allowed[origin] {
			http.Error(w, "CORS not allowed", http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/export-png", exportPngHandler)
	handler := corsHandler(mux)
	fmt.Println("Server working on http://localhost:8080")
	http.ListenAndServe(":8080", handler)
}
