package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func deploymentRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("token") != os.Getenv("KUBE_CI_WRITE_TOKEN") {
		fmt.Println("Rejected token")
		http.Error(w, "invalid token.", http.StatusUnauthorized)
		return
	}
	fmt.Println("---- Deployment ----")
	fmt.Println("Decoding PubSub message...")
	defer r.Body.Close()
	var message pubSubMessage
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Decoding Cloud Builder payload...")
	data, err := base64.StdEncoding.DecodeString(message.Message.Data)
	var status cloudBuilderStatus
	err = json.Unmarshal(data, &status)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Deploying to containers with image " + status.Images[0])
	b, e := builds[status.ID]
	if !e {
		b = &build{
			Created: time.Now().String(),
			History: make([]buildHistory, 0),
		}
		builds[status.ID] = b
	}
	b.History = append(b.History, buildHistory{
		Time:   time.Now().String(),
		Status: status.Status,
	})
	b.Status = status.Status
	if status.Status != "SUCCESS" {
		fmt.Println("Got status " + status.Status + " from " + status.Images[0])
		return
	}
	err = deploy(status)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type pubSubMessage struct {
	Message struct {
		Attributes struct {
			BuildID string `json:"buildId"`
			Status  string `json:"status"`
		} `json:"attributes"`
		Data      string `json:"data"`
		MessageID string `json:"message_id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

type cloudBuilderStatus struct {
	ID         string   `json:"id"`
	Status     string   `json:"status"`
	CreateTime string   `json:"createTime"`
	StartTime  string   `json:"startTime"`
	FinishTime string   `json:"finishTime"`
	Timeout    string   `json:"timeout"`
	Images     []string `json:"images"`
	LogsBucket string   `json:"logsBucket"`
	LogURL     string   `json:"logUrl"`
}
