# Serverless Event Gateway on Kubernetes

This guide walks you through provisioning a multi-node [Event Gateway](https://github.com/serverless/event-gateway) cluster on Kubernetes. This guide also demostrates how events can be routed across a diverse set of computing environments including Function as a Service (FaaS) offerings and containers running on Kubernetes. 

The [echo function](echo-function) and [echo application](echo) provide examples of how to handle HTTP events in the [Cloud Event](https://openevents.io) format leveraged by the Event Gateway.

## Tutorial

This tutorial assumes you have access to the [Google Cloud Platform](https://cloud.google.com) and have enabled the [Cloud Functions](https://cloud.google.com/functions) and [Kubernetes Engine](https://cloud.google.com/kubernetes-engine) APIs.

* [Creating a Kubernetes Cluster](#creating-a-kubernetes-cluster)
* [Bootstrapping an Event Gateway Cluster](#bootstrapping-an-event-gateway-cluster)
* [Routing Events to Google Cloud Functions](#routing-events-to-google-cloud-functions)
* [Routing Events to Kubernetes Services](#routing-events-to-kubernetes-services)
* [Clean Up](#clean-up)

## Creating a Kubernetes Cluster

The remainder of this tutorial requires access to a Kubernetes 1.9.7+ cluster. Google Cloud Platform users can create a Kubernetes cluster using the `gcloud` command: 

```
gcloud container clusters create event-gateway \
  --enable-autorepair \
  --cluster-version 1.9.7-gke.0 \
  --machine-type n1-standard-2 \
  --num-nodes 3 \
  --zone us-west1-c
```

## Bootstrapping an Event Gateway Cluster

In this section you will bootstrap a two node Event Gateway cluster suitable for learning and demonstration purposes.

> The Event Gateway configuration used in this tutorial is not recommended for production as it lacks any form of security or authentication.     

### Create an etcd Cluster

etcd is used to store and broadcast configuration across an Event Gateway cluster. A dedicated etcd cluster should be provisioned for the Event Gateway. Create the `etcd` statefulset:

```
kubectl apply -f statefulsets/etcd.yaml
```

```
statefulset "etcd" created
service "etcd" created
```

Verify the `etcd` cluster is up and running:

```
kubectl get pods
```
```
NAME      READY     STATUS    RESTARTS   AGE
etcd-0    1/1       Running   0          20s
```

### Create an Event Gateway Cluster

Create the `event-gateway` deployment:

```
kubectl apply -f deployments/event-gateway.yaml
```

```
deployment "event-gateway" created
service "event-gateway" created
```

At this point the Event Gateway should be deployed and exposed via an external load balancer accessible to external clients. Verify the Event Gateway is up and running:

```
kubectl get pods
```
```
NAME                             READY     STATUS    RESTARTS   AGE
etcd-0                           1/1       Running   0          1m
event-gateway-5ff8554766-r7ndx   1/1       Running   0          30s
event-gateway-5ff8554766-tp87g   1/1       Running   0          30s
```

Print the `event-gateway` service details:

```
kubectl get svc event-gateway
```
```
NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)                         AGE
event-gateway         LoadBalancer   10.15.248.210   XX.XXX.XXX.XX   4000:31061/TCP,4001:32247/TCP   11m
```

Extract the `event-gateway` service external IP address and store it:

```
EVENT_GATEWAY_IP=$(kubectl get svc \
  event-gateway \
  -o jsonpath={.status.loadBalancer.ingress[0].ip})
```

## Routing Events to Google Cloud Functions

In this section you will deploy the `echo` Google Cloud Function used to test the event routing functionality of the Event Gateway. Deploy the `echo` cloud function:

```
gcloud beta functions deploy echo \
  --source echo-function \
  --trigger-http
```

Get the HTTPS URL assigned to the `echo` cloud function and store it:

```
export FUNCTION_URL=$(gcloud beta functions describe echo \
  --format 'value(httpsTrigger.url)')
```

The `FUNCTION_URL` environment variable will be used in the next section to register the `echo` cloud function with the Event Gateway.

### Register the echo Goole Cloud Function

In this section you will register the `echo` cloud function with the Event Gateway.

Create a function registration request object:

```
cat > register-function.json <<EOF
{
  "functionId": "echo",
  "type": "http",
  "provider":{
    "url": "${FUNCTION_URL}"
  }
}
EOF
```

Register the `echo` cloud function by posting the function registration object to the Event Gateway:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data @register-function.json
```

At this point the `echo` cloud function has been registered with the Event Gateway, but before it can receive events a subscription must be created.

### Create a Subscription

A [subscription](https://github.com/serverless/event-gateway#subscriptions) binds an event to a function. In this section you will create an HTTP event subscription that binds the `echo` cloud function to HTTP events recieved on `POST` method and `/` path pair:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "echo",
    "event": "http",
    "method": "POST",
    "path": "/"
  }'
```

### Test the echo cloud function

With the `echo` cloud function registered and subscribed to HTTP events we can test the configuration by emitting HTTP events to the Event Gateway.

Submit an HTTP event to the Event Gateway:

```
curl -i --request POST \
  --url http://${EVENT_GATEWAY_IP}:4000/ \
  --data '{"message": "Hello world!"}'
```

The `echo` cloud function will respond with the data submitted in the HTTP event:

```
HTTP/1.1 200 OK
Compute-Type: function
Date: Tue, 08 May 2018 22:16:15 GMT
Content-Length: 27
Content-Type: text/plain; charset=utf-8

{"message": "Hello world!"}
```

> Notice the value of the `Compute-Type` HTTP header. It was set to `function` by the `echo` cloud function.

Review the `echo` cloud function logs:

```
gcloud beta functions logs read echo
```

```
LEVEL  NAME  EXECUTION_ID  TIME_UTC                 LOG
D      echo  4uczimni6d70  2018-05-08 23:24:11.206  Function execution started
I      echo  4uczimni6d70  2018-05-08 23:24:11.458  Handling HTTP event 13e2cfa2-3c86-42dc-a8be-a01648b6444c
D      echo  4uczimni6d70  2018-05-08 23:24:11.546  Function execution took 341 ms, finished with status code: 200
```

## Routing Events to Kubernetes Services

In most Serverless deployments events are typically routed to functions running on a hosted FaaS platform such as [Google Cloud Functions](https://cloud.google.com/functions) or [AWS Lambda](https://aws.amazon.com/lambda). However this is not a hard requirement. Events can be routed to any HTTP endpoint, including endpoints backed by containers running on Kubernetes.

In this section you will deploy the `gcr.io/hightowerlabs/echo:event-gateway` container using Kubernetes and route HTTP events to it.

Create the `echo` deployment and service:

```
kubectl create -f deployments/echo.yaml
```

```
deployment "echo" created
service "echo" created
```

Verify the `echo` deployment is up and running:

```
kubectl get pods
```
```
NAME                             READY     STATUS    RESTARTS   AGE
echo-77d48cb484-2h5cl            1/1       Running   0          30s
etcd-0                           1/1       Running   0          2m
event-gateway-5ff8554766-r7ndx   1/1       Running   0          1m
event-gateway-5ff8554766-tp87g   1/1       Running   0          1m
```

Register the `echo` service with the Event Gateway:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "echo-service",
    "type": "http",
    "provider":{
      "url": "http://echo.default.svc.cluster.local"
    }
  }'
```

> The HTTP URL follows the standard format for accessing cluster local services. In this case the `echo` deployment runs in the `default` namespace. This configuration works because the Event Gateway is running in the same cluster as the `echo` deployment.

At this point the `echo` service has been registered with the Event Gateway, but before it can receive events a subscription must be created. There can only be one binding for HTTP events mapped to a specific path and method pair.

Delete the current subscription for HTTP events on the `POST` method `/` path pair:

```
curl -X DELETE \
  http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/subscriptions/http,POST,%2F
```

Next, create an HTTP event subscription for the `echo-service` function:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "echo-service",
    "event": "http",
    "method": "POST",
    "path": "/"
  }'
```

Test the `echo` service by emitting an HTTP event to the Event Gateway:

```
curl -i --request POST \
  --url http://${EVENT_GATEWAY_IP}:4000/ \
  --data '{"message": "Hello World!"}'
```

```
HTTP/1.1 200 OK
Compute-Type: container
Date: Tue, 08 May 2018 23:35:35 GMT
Content-Length: 27
Content-Type: text/plain; charset=utf-8

{"message": "Hello World!"}
```

> Notice the value of the `Compute-Type` HTTP header. It was set to `container` by the echo service.

Review the `echo` container logs:

```
kubectl logs echo-77d48cb484-2h5cl
```

```
2018/05/08 23:30:04 Starting HTTP server...
2018/05/08 23:35:35 Handling HTTP event f3a37c57-d85a-4942-b92c-cef56713d538 ...
```

## Clean Up

```
kubectl delete -f deployments
```

```
kubectl delete -f statefulsets
```

```
gcloud container clusters delete event-gateway --quiet
```
