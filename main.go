package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title HTTP Proxy Server API
// @version 1.0
// @description This is a sample server for proxying HTTP requests.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Length  int               `json:"length"`
}

var requestStore sync.Map

// handler handles incoming proxy requests
// @Summary Handle proxy requests
// @Description Handle incoming HTTP requests and forward them to the specified URL
// @Accept  json
// @Produce  json
// @Param   request body Request true "Proxy Request"
// @Success 200 {object} Response
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Router / [post]
func handler(w http.ResponseWriter, r *http.Request) {
	var req Request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Method == "" || req.URL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest(req.Method, req.URL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	headers := make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = v[0]
	}

	id := fmt.Sprintf("%v", r.Header.Get("X-Request-ID"))
	response := Response{
		ID:      id,
		Status:  resp.StatusCode,
		Headers: headers,
		Length:  int(resp.ContentLength),
	}

	requestStore.Store(id, response)

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// healthCheckHandler provides a simple health check endpoint
// @Summary Health check
// @Description Check if the server is running
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /health [get]
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
