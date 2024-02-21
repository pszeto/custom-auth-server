package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	hostname, _ := os.Hostname()
	log.Printf("Handling %s request : %s %s %s %s headers(%s)\n", req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)

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
	log.Printf("Handling %s request : %s %s %s %s headers(%s)\n", req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
	w.Header().Add("x-auth-server", hostname)
	id := uuid.New()
	w.Header().Add("etag", id.String())
	w.Header().Add("x-e-tag", id.String())
	w.Header().Add("content-type", "application/json")
	w.Header().Add("transfer-encoding", "chunked")

	resp := make(map[string]string)
	resp["message"] = "success"
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

func delay(w http.ResponseWriter, req *http.Request) {
	hostname, _ := os.Hostname()
	log.Printf("Handling %s request : %s %s %s %s headers(%s)\n", req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
	w.Header().Add("x-auth-server", hostname)
	w.Header().Add("content-type", "application/json")
	path := req.URL.Path
	pathSplit := strings.Split(path, "/")
	delay := pathSplit[2]
	delayAmount, err := strconv.Atoi(delay)
	if err != nil {
		delayAmount = 1

	}
	log.Printf("Delaying response %s seconds.", delay)
	time.Sleep(time.Duration(delayAmount) * time.Second)
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
	http.HandleFunc("/delay/1", delay)
	http.HandleFunc("/delay/2", delay)
	http.HandleFunc("/delay/3", delay)
	http.HandleFunc("/delay/4", delay)
	http.HandleFunc("/delay/5", delay)
	http.HandleFunc("/delay/6", delay)
	http.HandleFunc("/delay/7", delay)
	http.HandleFunc("/delay/8", delay)
	http.HandleFunc("/delay/9", delay)
	http.HandleFunc("/delay/10", delay)

	log.Println("Starting server on port" + port)
	log.Println("Version 1.14")
	http.ListenAndServe(port, nil)
}
