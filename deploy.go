package main

import (
	"context"
	"fmt"
	"strings"
)

func deploy(status cloudBuilderStatus) error {
	fmt.Println("DEPLOYING " + status.Images[0])
	client, err := kubeConnect()
	if err != nil {
		return err
	}
	image := strings.Split(status.Images[0], ":")[0]
	fmt.Println("Scanning cluster for namespaces...")
	ctx := context.Background()
	namespaces, err := client.CoreV1().ListNamespaces(ctx)
	for _, namespace := range namespaces.Items {
		fmt.Println("Scanning cluster for deployments...")
		deployments, _ := client.ExtensionsV1Beta1().ListDeployments(ctx, namespace.Metadata.GetName())
		for _, deployment := range deployments.Items {
			_, ci := deployment.GetMetadata().GetAnnotations()["kube-ci"]
			fmt.Println("Skipped deployment " + deployment.GetMetadata().GetName())
			if !ci {
				continue
			}
			fmt.Println("Deploying on " + deployment.GetMetadata().GetName())
			for _, container := range deployment.Spec.Template.Spec.Containers {
				if strings.HasPrefix(container.GetImage(), image) {
					fmt.Println("Setting container " + container.GetName() + " image")
					container.Image = &status.Images[0]
				}
			}
			fmt.Println("Updating deployment " + deployment.GetMetadata().GetName())
			client.ExtensionsV1Beta1().UpdateDeployment(ctx, deployment)
		}
	}
	return nil
}
