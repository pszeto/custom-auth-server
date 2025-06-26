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
	logId := uuid.New()
	hostname, _ := os.Hostname()
	log.Printf("[%s] Handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
	headers := req.Header
	log.Println("[%s] Request Headers:", logId)
	for name, values := range headers {
		for _, value := range values {
			log.Printf("[%s] %s: %s\n", logId, name, value)
		}
	}
	
	log.Println("[%s] Request Headers End", logId)
	apiKey, ok := os.LookupEnv("API_KEY")
	w.Header().Set("X-Server", "custom-auth-server")
	if !ok {
		log.Println("[%s] API_KEY is not present. Setting to default : 828c3c5f-30ab-4291-8ad1-7cc33ba0be4f", logId)
		apiKey = "828c3c5f-30ab-4291-8ad1-7cc33ba0be4f"
	} else {
		log.Printf("[%s] API_KEY: %s\n", logId, apiKey)
	}

	
	
	incomingApiKey, ok := headers["X-Api-Key"]

	w.Header().Add("x-auth-server", hostname)
	if ok {
		log.Printf("[%s] x-api-key %s\n", logId, incomingApiKey)
		if incomingApiKey[0] == apiKey {
			log.Printf("[%s] - %s [200] %s %s \n", logId, hostname, req.Host, req.URL.Path)
			log.Printf("[%s] Completed handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
			w.WriteHeader(200)
			w.Write([]byte("Success"))
		} else {
			log.Printf("[%s], %s [401] %s %s Reason: Failed Authorization - keys don't match. API_KEY: %s x-api-key %s \n", logId, hostname, req.Host, req.URL.Path, apiKey, incomingApiKey[0])
			log.Printf("[%s] Completed handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
			w.WriteHeader(401)
			w.Write([]byte("Not authorized"))
		}
	} else {
		log.Printf("[%s] %s [401] %s %s Reason: Failed Authorization - x-api-key header is not present\n", logId, hostname, req.Host, req.URL.Path)
		log.Printf("[%s] Completed handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
		w.WriteHeader(401)
		w.Write([]byte("Not authorized"))
	}
}

func noauth(w http.ResponseWriter, req *http.Request) {
	hostname, _ := os.Hostname()
	logId := uuid.New()
	log.Printf("[%s] Handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
	w.Header().Add("x-auth-server", hostname)
	id := uuid.New()
	w.Header().Add("etag", id.String())
	w.Header().Add("x-e-tag", id.String())
	w.Header().Add("content-type", "application/json")
	w.Header().Add("transfer-encoding", "chunked")

	resp := make(map[string]string)
	resp["message"] = "success"
	jsonResp, _ := json.Marshal(resp)
	log.Printf("[%s] Completed handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
	w.Write(jsonResp)
}

func delay(w http.ResponseWriter, req *http.Request) {
	logId := uuid.New()
	hostname, _ := os.Hostname()
	log.Printf("[%s] Handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
	w.Header().Add("x-auth-server", hostname)
	w.Header().Add("content-type", "application/json")
	path := req.URL.Path
	log.Printf("[%s] Path %s value in url\n", logId, path)
	pathSplit := strings.Split(path, "/")
	log.Printf("[%s] pathSplit %s value in url\n", logId, pathSplit)
	delay := pathSplit[2]
	log.Printf("[%s] Delay %s value in url\n", logId, delay)
	delayAmount, err := strconv.Atoi(delay)
	if err != nil {
		delayAmount = 1

	}
	log.Printf("[%s] Delaying response %s seconds.", logId, delay)
	time.Sleep(time.Duration(delayAmount) * time.Second)
	log.Printf("[%s] Finished Delaying response %s seconds.", logId, delay)
	resp := make(map[string]string)
	resp["message"] = "success"
	jsonResp, _ := json.Marshal(resp)
	log.Printf("[%s] Completed handling %s request : %s %s %s %s headers(%s)\n", logId, req.Proto, hostname, req.Host, req.Method, req.URL.Path, req.Header)
	w.Write(jsonResp)
}

func Run(addr string, sslAddr string, ssl map[string]string) chan error {

	errs := make(chan error)

	// Starting HTTP server
	go func() {
		log.Printf("Staring HTTP service on %s", addr)

		if err := http.ListenAndServe(addr, nil); err != nil {
			errs <- err
		}

	}()

	// Starting HTTPS server
	go func() {
		log.Printf("Staring HTTPS service on %s", sslAddr)
		if err := http.ListenAndServeTLS(sslAddr, ssl["cert"], ssl["key"], nil); err != nil {
			errs <- err
		}
	}()

	return errs
}

func main() {
	httpPort, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Println("HTTP_PORT not defined.  Defaulting to 8080")
		httpPort = ":8080"
	} else {
		httpPort = ":" + httpPort
	}

	httpsPort, ok := os.LookupEnv("HTTPS_PORT")
	if !ok {
		log.Println("HTTPS_PORT not defined.  Defaulting to 8443")
		httpsPort = ":8443"
	} else {
		httpsPort = ":" + httpsPort
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

	http.HandleFunc("/delay/", delay)

	log.Println("Version 1.15")

	errs := Run(httpPort, httpsPort, map[string]string{
		"cert": "server.crt",
		"key":  "server.key",
	})

	// This will run forever until channel receives error
	select {
	case err := <-errs:
		log.Printf("Could not start serving service due to (error: %s)", err)
	}
}
