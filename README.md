# Serverless Event Gateway on Kubernetes

This guide will walk you through provisioning a multi-node [Event Gateway](https://github.com/serverless/event-gateway) cluster on Kubernetes. The goal of this guide is to introduce you to the Event Gateway and get a feel for the role it plays in a Serverless Architecture.

This guide will also demostrate how events can be routed across a diverse set of computing environments including Function as a Service (FaaS) offerings and containers running on Kubernetes. 

## Tutorial

* [Creating a Kubernetes Cluster](#creating-a-kubernetes-cluster)
* [Bootstrapping an Event Gateway Cluster](#bootstrapping-an-event-gateway-cluster)
* [Routing Events to Google Cloud Functions](#routing-events-to-google-cloud-functions)
* [Routing Events to Kubernetes Services](#routing-events-to-kubernetes-services)


### Creating a Kubernetes Cluster

This tutorial assumes you have access to a Kubernetes 1.9.6+ cluster and [Google Cloud Functions](https://cloud.google.com/functions)*.

```
gcloud container clusters create event-gateway \
  --async \
  --enable-autorepair \
  --enable-network-policy \
  --cluster-version 1.9.7-gke.0 \
  --machine-type n1-standard-2 \
  --num-nodes 3 \
  --zone us-west1-c
```

> Any backend that can respond to HTTP request will work. Google Cloud Functions is only being used to streamline the learning process.

### Bootstrapping an Event Gateway Cluster

In this section you will create a two node Event Gateway cluster backed by a single node etcd cluster. This deployment is only suitable for learning and demonstration purposes. This configuration is not recommend for production.

#### Create an etcd Cluster

etcd is used to store and broadcast configuration across an Event Gateway cluster.

Create the `etcd` statefulset:

```
kubectl apply -f etcd.yaml
```

Verify `etcd` is up and running:

```
kubectl get pods
```
```
NAME      READY     STATUS    RESTARTS   AGE
etcd-0    1/1       Running   0          50s
```

#### Create an Event Gateway Cluster

Create the `event-gateway` deployment:

```
kubectl apply -f event-gateway.yaml
```

At this point the Event Gateway should be up and running and exposed via an external loadbalancer.

```
kubectl get pods
```
```
NAME                             READY     STATUS    RESTARTS   AGE
etcd-0                           1/1       Running   0          2m
event-gateway-5ff8554766-9x2zx   1/1       Running   0          15s
event-gateway-5ff8554766-kqwwg   1/1       Running   0          15s
```

Print the `event-gateway` service details:

```
kubectl get svc event-gateway
```
```
NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)                         AGE
event-gateway         LoadBalancer   10.15.248.210   XX.XXX.XXX.XX   4000:31061/TCP,4001:32247/TCP   11m
```

> In this tutorial the Event Gateway is not protected by TLS or authentication and should only be used for learning the basics.

Get the external IP address assigned to the `event-gateway` service and store it:

```
EVENT_GATEWAY_IP=$(kubectl get svc \
  event-gateway \
  -o jsonpath={.status.loadBalancer.ingress[0].ip})
```

### Routing Events to Google Cloud Functions

In this section you will write and deploy a Google Cloud Function which will be used to test the event routing functionality of the Event Gateway cluster.

Create the `helloworld` function:

```
cat > index.js <<EOF
exports.helloworld = (req, res) => {
  res.status(200).send('Success: ' + req.body.data.message);
};
EOF
```

Deploy the `helloworld` function:

```
gcloud beta functions deploy helloworld --trigger-http
```

Get the HTTPS URL assigned to the `helloworld` function and store it:

```
export FUNCTION_URL=$(gcloud beta functions describe helloworld \
  --format 'value(httpsTrigger.url)')
```

#### Register the Helloworld Goole Cloud Function

In this section you will register the `helloworld` function with the Event Gateway.

Register the `helloworld` function. Create the function registration request body:

```
cat > register-function.json <<EOF
{
  "functionId": "helloworld",
  "type": "http",
  "provider":{
    "url": "${FUNCTION_URL}"
  }
}
EOF
```

Post the function registration to the Event Gateway:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data @register-function.json
```

At this point the `helloworld` cloud function has been registered with the Event Gateway, but before it can receive events a subscription must be created.

#### Create a Subscription

A subscription binds an event to function. Multiple subscriptions can be used to broadcast a single event across multiple functions.

Post an event subscription to the Event Gateway which binds the `helloworld` function to a custom event named `test.event`:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "helloworld",
    "event": "test.event",
    "path": "/"
  }'
```

#### Emit an event

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4000/ \
  --header 'content-type: application/json' \
  --header 'event: test.event' \
  --data '{"message": "Hello!"}'
```

Review the Cloud Functions logs:

```
gcloud beta functions logs read helloworld
```

### Routing Events to Kubernetes Services

```
kubectl apply -f helloworld.yaml
```

```
kubectl get pods
```
```
NAME                             READY     STATUS    RESTARTS   AGE
etcd-0                           1/1       Running   0          3m
event-gateway-5ff8554766-9x2zx   1/1       Running   0          1m
event-gateway-5ff8554766-kqwwg   1/1       Running   0          1m
helloworld-749c6b5df7-d24qh      1/1       Running   0          1m
```

Register the `helloworld` container. Create the function registration request body:

```
cat > register-container.json <<EOF
{
  "functionId": "helloworld-container",
  "type": "http",
  "provider":{
    "url": "http://helloworld.default.svc.cluster.local"
  }
}
EOF
```

Post the function registration to the Event Gateway:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data @register-container.json
```

At this point the `helloworld` Pod has been registered with the Event Gateway, but before it can receive events a subscription must be created.

### Create a Subscription

A subscription binds an event to function. Multiple subscriptions can be used to broadcast a single event across multiple functions.

Post an event subscription to the Event Gateway which binds the `helloworld` Pod to a custom event named `test.event`:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "helloworld-container",
    "event": "test.event",
    "path": "/"
  }'
```

### Emit an event

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4000/ \
  --header 'content-type: application/json' \
  --header 'event: test.event' \
  --data '{"message": "Hello!"}'
```

Review the `helloworld` Pod logs:

```
kubectl logs helloworld-749c6b5df7-d24qh
```

```
2018/05/08 15:54:09 Starting HTTP server...
2018/05/08 15:54:18 Handling event a8b15854-0aa8-483d-979f-0c73d36173bd from https://serverless.com/event-gateway/#transformationVersion=0.1 ...
map[message:Hello!]
```
