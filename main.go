package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
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
	if ok {
		log.Printf("x-api-key %s\n", incomingApiKey)
		if incomingApiKey[0] == apiKey {
			w.WriteHeader(200)
			log.Println("Successfully Authorized")
			w.Write([]byte("Success"))
		} else {
			log.Println("Failed Authorization: keys don't match")
			log.Printf("API_KEY: %s\n", apiKey)
			log.Printf("x-api-key: %s\n", incomingApiKey[0])
			w.WriteHeader(401)
			w.Write([]byte("Not authorized"))
		}
	} else {
		log.Println("x-api-key header is not present")
		w.WriteHeader(401)
		w.Write([]byte("Not authorized"))
	}
}

func noauth(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Success"))
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
	http.ListenAndServe(port, nil)
}
