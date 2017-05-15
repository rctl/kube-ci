package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var builds map[string]*build

func main() {
	builds = make(map[string]*build)
	http.HandleFunc("/deployments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != os.Getenv("KUBE_CI_WRITE_TOKEN") && r.URL.Query().Get("token") != os.Getenv("KUBE_CI_READ_TOKEN") {
			http.Error(w, "invalid token.", http.StatusUnauthorized)
			return
		}
		list := make([]*build, 0)
		for _, b := range builds {
			list = append(list, b)
		}
		json.NewEncoder(w).Encode(list)
	})
	http.HandleFunc("/deploy", deploymentRequest)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	fmt.Println("Ready")
	fmt.Print(http.ListenAndServe(":80", http.DefaultServeMux))
}

type build struct {
	Status  string `json:"status"`
	Image   string `json:"image"`
	Created string `json:"created"`
	History []buildHistory
}

type buildHistory struct {
	Time   string `json:"time"`
	Status string `json:"status"`
}
