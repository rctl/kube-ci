package main

import "github.com/ericchiang/k8s"

func kubeConnect() (*k8s.Client, error) {
	client, err := k8s.NewInClusterClient()
	return client, err
}
