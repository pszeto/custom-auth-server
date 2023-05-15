package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

var startTime time.Time

func uptime() time.Duration {
	return time.Since(startTime)
}

func init() {
	startTime = time.Now()
}

func status(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server", "custom-auth-server")
	resp := make(map[string]string)
	resp["uptime"] = uptime().String()
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	return
}

func auth(w http.ResponseWriter, req *http.Request) {
	apiKey, ok := os.LookupEnv("API_KEY")
	w.Header().Set("X-Server", "custom-auth-server")
	if !ok {
		log.Println("API_KEY is not present. Setting to default : 828c3c5f-30ab-4291-8ad1-7cc33ba0be4f")
		apiKey = "828c3c5f-30ab-4291-8ad1-7cc33ba0be4f"
	} else {
		log.Printf("API_KEY: %s\n", apiKey)
	}

	headers := req.Header
	incomingApiKey, ok := headers["X-Api-Key"]
	hostname, _ := os.Hostname()
	w.Header().Add("x-auth-server", hostname)
	if ok {
		log.Printf("x-api-key %s\n", incomingApiKey)
		if incomingApiKey[0] == apiKey {
			log.Printf("%s [200] %s %s \n", hostname, req.Host, req.URL.Path)
			w.WriteHeader(200)
			w.Write([]byte("Success"))
		} else {
			log.Printf("%s [401] %s %s Reason: Failed Authorization - keys don't match. API_KEY: %s x-api-key %s \n", hostname, req.Host, req.URL.Path, apiKey, incomingApiKey[0])
			w.WriteHeader(401)
			w.Write([]byte("Not authorized"))
		}
	} else {
		log.Printf("%s [401] %s %s Reason: Failed Authorization - x-api-key header is not present\n", hostname, req.Host, req.URL.Path)
		w.WriteHeader(401)
		w.Write([]byte("Not authorized"))
	}
}

func noauth(w http.ResponseWriter, req *http.Request) {
	hostname, _ := os.Hostname()
	log.Printf("%s [200] %s %s headers(%s)\n", hostname, req.Host, req.URL.Path, req.Header)
	w.Header().Add("x-auth-server", hostname)
	id := uuid.New()
	w.Header().Add("etag", id.String())
	w.Header().Add("x-e-tag", id.String())
	resp := make(map[string]string)
	resp["message"] = "success"
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		log.Println("Port not defined.  Defaulting to 8000")
		port = ":8000"
	} else {
		port = ":" + port
	}

	serverType, ok := os.LookupEnv("SERVER_TYPE")
	if !ok {
		log.Println("SERVER_TYPE not defined.  Defaulting to generic")
		serverType = "GENERIC"
	} else {
		if serverType == "AUTH" {
			log.Println("SERVER_TYPE is AUTH")
		} else {
			serverType = "GENERIC"
		}

	}
	http.HandleFunc("/status", status)
	if serverType == "AUTH" {
		http.HandleFunc("/", auth)
	} else {
		http.HandleFunc("/", noauth)
	}

	log.Println("Starting server on port" + port)
	log.Println("Version 1.13")
	http.ListenAndServe(port, nil)
}
