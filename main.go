package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/", proxyHandler)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Get the X-Method-Override header
	method := r.Header.Get("X-Method-Override")
	if method == "" {
		// If the header is not present, use the default method (POST in this case)
		method = "POST"
	}

	// Get the X-URL header
	targetURL := r.Header.Get("X-URL")
	if targetURL == "" {
		http.Error(w, "X-URL header is required", http.StatusBadRequest)
		return
	}

	// Create a new request to the target URL with the specified method
	req, err := http.NewRequest(method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Forward the headers from the original request to the new request
	for header, values := range r.Header {
		// Skip the X-Method-Override and X-URL headers
		if header == "X-Method-Override" || header == "X-URL" {
			continue
		}
		// Set the header in the new request
		for _, value := range values {
			req.Header.Add(header, value)
		}
	}

	// Make the request to the target URL
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read and forward the response from the target URL
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Forward the status code and headers from the response
	w.WriteHeader(resp.StatusCode)
	for header, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	// Write the response body
	fmt.Fprint(w, string(responseBody))
}
