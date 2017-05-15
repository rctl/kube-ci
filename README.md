## kube-ci - Continuous delivery for Kubernetes

*kube-ci* is a project to bring simple continuous delivery to Kubernetes without any custom configurations to enable for new deployments.

kube-ci operates inside your Kubernetes cluster and listen for deplyment triggers. When a trigger is received kube-ci will automatically update your images for deployments where kube-ci is enabled. It has built in support for Google Cloud Builds and will automatically parse build updates received by Cloud Builds through Pub/Sub.

**Sample configuration**

``` yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-app
  annotations:
    kube-ci: "true"
...
```
Annotate your Kubernetes Deployment with *kube-ci: "true"* to enable automatic updates.

### Google Container Engine

kube-ci works best with GKE but can be customized to work with any Kubernetes or build setups. The following steps will guide you through setting up Google Cloud Builder with GitHub and setup kube-ci with your GKE cluster.

**Configure kube-ci deployment**

Update *kube-ci.yaml* to reflect your own settings.

- Change *my-kube-ci-domain.example.com* to a domain name that you can use to serve kube-ci in your cluster.
- Recommended: Enable TLS for your Kubernetes Ingress by editing the Ingress configuration.
- Change *my-read-token* to a secure, random, key that will be used for read access in kube-ci.
- Change *my-write-token* to a secure, random, key that will be used for write access in kube-ci.

`kubectl apply -f kube-ci.yaml`

**Setup Pub/Sub**

First you need to verify your kube-ci domain to be able receive Pub/Sub updated on your Cloud Builds.
Go into API Manager -> Credentials -> Add domain in Google Cloud Console.
Add the domain you setup in the previous step.

![Google Cloud setup](https://raw.githubusercontent.com/rctl/kube-ci/master/images/verify-domain.png)

Go into Pub/Sub -> Topics -> projects/your-project/topics/cloud-builds -> Create Subscription
Create a Pub/Sub subscription for kube-ci, use the same domain and write key you used in the previous step.

![Google Cloud setup](https://raw.githubusercontent.com/rctl/kube-ci/master/images/pub-sub.png)

**Setup a build trigger**

Go into Cloud Build -> Build Triggers -> Add Trigger and select a repo and setting to use for your build trigger.

Example:

![Google Cloud setup](https://raw.githubusercontent.com/rctl/kube-ci/master/images/trigger.png)

You trigger will automatically build your Docker Container when you push to a specific branch or tag.
This will then update kube-ci through Pub/Sub which will update your Kubernetes Deployments.

**Create your Kubernetes Deployment and enable kube-ci**

Create (or edit and existing) deployment in Kubernetes that uses the image you created your build trigger for.

Example:

``` yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: my-app
  annotations:
    kube-ci: "true"
spec:
  replicas: 2
  strategy:
    rollingUpdate:
      maxSurge: 0
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: web
        image: gcr.io/my-project/my-repo:a-tag
        readinessProbe:
          httpGet:
            path: /
            port: 80
          periodSeconds: 5
          successThreshold: 2
          failureThreshold: 2
        ports:
        - containerPort: 80
```

**Push to master**

Push to your repository and observe your deployment being automatically updated when your Cloud Build finishes.

**Get statuses**

You can use your kube-ci read token to get statueses for your builds:

`curl https://my-kube-ci-domain.example.com/deployments?token=my-read-token`

### License

Project is licenced under MIT.