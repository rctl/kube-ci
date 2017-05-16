package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"io/ioutil"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func subscribe() {
	ctx := context.Background()
	project := os.Getenv("GCP_PROJECT")
	serviceAccountPath := "google_service_account.json"
	var client *pubsub.Client
	if _, err := os.Stat(serviceAccountPath); !os.IsNotExist(err) {
		fmt.Println("Using service account from file")
		//Use service account from file
		tokenData, err := ioutil.ReadFile(serviceAccountPath)
		if err != nil {
			panic(err.Error())
		}
		var serviceAccountData googleServiceAccount
		err = json.Unmarshal(tokenData, &serviceAccountData)
		if err != nil {
			panic(err.Error())
		}
		project = serviceAccountData.ProjectID
		client, err = pubsub.NewClient(ctx, project, option.WithServiceAccountFile(serviceAccountPath))
		if err != nil {
			fmt.Println(err)
			panic("Unable to initialize Pub/Sub client.")
		}
	} else {
		fmt.Println("Using service account from kubernetes")
		credentials, err := google.FindDefaultCredentials(ctx, pubsub.ScopePubSub)
		if err != nil || credentials.ProjectID == "" {
			panic("Project ID could not be automatcially found, please set the GCP_PROJECT env. vairable.")
		}
		project = credentials.ProjectID
		client, err = pubsub.NewClient(ctx, project)
		if err != nil {
			fmt.Println(err)
			panic("Unable to initialize Pub/Sub client.")
		}
	}
	fmt.Println("GCP project found: " + project)
	topic := client.Topic("cloud-builds")
	subs := topic.Subscriptions(ctx)
	var subscription *pubsub.Subscription
	for {
		sub, err := subs.Next()
		if err != nil {
			break
		}
		if sub.ID() == "kube-ci" {
			subscription = sub
		}
	}
	if subscription == nil {
		fmt.Println("No Pub/Sub subscription found, creating new...")
		var err error
		subscription, err = client.CreateSubscription(ctx, "kube-ci", topic, time.Second*30, nil)
		if err != nil {
			fmt.Println(err)
			panic("Unable to initialize Pub/Sub client.")
		}
	}
	go subscription.Receive(ctx, func(context context.Context, message *pubsub.Message) {
		fmt.Println("Received Pub/Sub message")
		message.Ack()
		var status cloudBuilderStatus
		err := json.Unmarshal(message.Data, &status)
		if err != nil {
			fmt.Println("Could not decode message")
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
			return
		}
	})
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

type googleServiceAccount struct {
	ProjectID string `json:"project_id"`
}
