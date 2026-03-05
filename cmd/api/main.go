package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type RawData struct {
	Data     []RequestData `json:"data"`
	Title    string        `json:"title"`
	DateFrom string        `json:"dateFrom"`
	DateTo   string        `json:"dateTo"`
}

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
	var rawData RawData
	err := json.NewDecoder(r.Body).Decode(&rawData)
	if err != nil {
		fmt.Println("Error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	values := make(plotter.Values, 0, len(rawData.Data))
	labels := make([]string, 0, len(rawData.Data))
	for _, value := range rawData.Data {
		values = append(values, float64(value.N_Reports))
		labels = append(labels, value.Name)
	}
	p := plot.New()
	p.X.Tick.Label.Rotation = 0.5
	p.X.Tick.Label.XAlign = draw.XRight
	p.X.Tick.Label.YAlign = draw.YCenter
	p.X.Padding = vg.Points(10)
	p.Y.Padding = vg.Points(10)
	p.X.Tick.Label.Font.Size = vg.Points(10)
	title := fmt.Sprintf("%s\n(%s - %s)", rawData.Title, rawData.DateFrom, rawData.DateTo)
	p.Title.Text = title
	p.Y.Label.Text = "Número de reportes"
	bar, err := plotter.NewBarChart(values, vg.Points(30))
	if err != nil {
		log.Fatal(err)
	}
	bar.Color = color.RGBA{R: 0, G: 192, B: 255, A: 255}
	bar.LineStyle.Width = vg.Length(0)
	p.Add(bar)
	p.NominalX(labels...)
	var buf bytes.Buffer
	wt, err := p.WriterTo(10*vg.Inch, 6*vg.Inch, "png")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="reporte.png"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, errWrite := wt.WriteTo(&buf)
	if errWrite != nil {
		log.Fatal(errWrite)
	}
	w.Write(buf.Bytes())
	return
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
