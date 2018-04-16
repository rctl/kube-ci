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

> **Note!** This project has not been used in production by the creator (rctl) since end of 2017. Therefore I would strongly advice to do thurough testing before using it in a production environment. Feel free to leave any issues for feedback. The project is still maintained on request.

### Setup with Google Container Engine

kube-ci works best with GKE but can be customized to work with any Kubernetes or build setups. The following steps will guide you through setting up Google Cloud Builder with GitHub and setup kube-ci with your GKE cluster. Your GKE cluster need read access to Pub/Sub for kube-ci to work properly.

**Deploy kube-ci in Kubernetes**

Run `kubectl apply -f kube-ci.yaml` to deploy kube-ci.

**Setup a build trigger**

Go into Cloud Build -> Build Triggers -> Add Trigger and select a repo and setting to use for your build trigger.

Example:

![Google Cloud setup](https://raw.githubusercontent.com/rctl/kube-ci/master/images/trigger.png)

You trigger will automatically build your Docker Container when you push to a specific branch or tag.
This will then update kube-ci through Pub/Sub which will update your Kubernetes Deployments.

**Create or update your Kubernetes Deployment to enable kube-ci**

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

### kube-ci HTTP API

kube-ci has an HTTP API that can be used to fetch build statueses.
**Get statuses**

Access build statueses from inside the cluster with:

`curl http://kube-ci/`

If you have set a read token use the following:

`curl http://kube-ci/?token=my-read-token`

**From the Internet**

It is safe to expose kube-ci to the internet. If kube-ci is exposed to the internet it is recommended to use a read token and TLS. Configure this with your own ingress controller.

### Custom Service Account

**If your cluster does not have access to Pub/Sub by default** you can enable access by setting up kube-ci with a service account secret.
Create your service account in the cloud config (with access to Pub/Sub scopes) and use the following command to add it to kubernetes `kubectl create secret generic pubsub-service-account --from-file=google_service_account.json`.

To configure kube-ci to access Pub/Sub with the service account from your kubernetes secret with this yaml config:

``` yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-ci
spec:
  template:
    metadata:
      labels:
        app: kube-ci
    spec:
      containers:
      - name: kube-ci
        image: rctl/kube-ci:1.1.5-alpha
        ports:
        - containerPort: 80
        volumeMounts:
        - name: serviceaccount
          mountPath: "/go/src/app"
          readOnly: true
      volumes:
      - name: serviceaccount
        secret:
          secretName: "pubsub-service-account"
```

### License

Project is licenced under MIT.
