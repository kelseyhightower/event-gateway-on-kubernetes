# Serverless Event Gateway on Kubernetes

This guide will walk you through provisioning a multi-node [Event Gateway](https://github.com/serverless/event-gateway) cluster on Kubernetes. The goal of this guide is to introduce you to the Event Gateway and get a feel for the role it plays in a Serverless Architecture.

This guide will also demostrate how events can be routed across a diverse set of computing environments including Function as a Service (FaaS) offerings and containers running on Kubernetes. 

## Tutorial

* [Creating a Kubernetes Cluster](#creating-a-kubernetes-cluster)
* [Bootstrapping an Event Gateway Cluster](#bootstrapping-an-event-gateway-cluster)
* [Routing Events to Google Cloud Functions](#routing-events-to-google-cloud-functions)
* [Routing Events to Kubernetes Services](#routing-events-to-kubernetes-services)

## Creating a Kubernetes Cluster

This tutorial assumes you have access to a Kubernetes 1.9.7+ cluster. If you are using the Google Cloud Platform you can create a Kubernetes cluster using the `gcloud` command: 

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

## Bootstrapping an Event Gateway Cluster

In this section you will create a two node Event Gateway cluster backed by a single node etcd cluster. This deployment is only suitable for learning and demonstration purposes. This configuration is not recommend for production.

### Create an etcd Cluster

etcd is used to store and broadcast configuration across an Event Gateway cluster.

Create the `etcd` statefulset:

```
kubectl apply -f statefulsets/etcd.yaml
```

Verify `etcd` is up and running:

```
kubectl get pods
```
```
NAME      READY     STATUS    RESTARTS   AGE
etcd-0    1/1       Running   0          50s
```

### Create an Event Gateway Cluster

Create the `event-gateway` deployment:

```
kubectl apply -f deployments/event-gateway.yaml
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

## Routing Events to Google Cloud Functions

In this section you will write and deploy a Google Cloud Function which will be used to test the event routing functionality of the Event Gateway cluster.

Deploy the `echo` function:

```
gcloud beta functions deploy echo \
  --source echo-function \
  --trigger-http
```

Get the HTTPS URL assigned to the `echo` function and store it:

```
export FUNCTION_URL=$(gcloud beta functions describe echo \
  --format 'value(httpsTrigger.url)')
```

### Register the echo Goole Cloud Function

In this section you will register the `echo` function with the Event Gateway.

Register the `echo` function. Create the function registration request body:

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

Post the function registration to the Event Gateway:

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data @register-function.json
```

At this point the `echo` cloud function has been registered with the Event Gateway, but before it can receive events a subscription must be created.

### Create a Subscription

A subscription binds an event to a function. Create an HTTP event subscription which binds the `echo` function to a HTTP event on the `/` path and the `POST` method:

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

### Emit an HTTP Event

```
curl -i --request POST \
  --url http://${EVENT_GATEWAY_IP}:4000/ \
  --data '{"message": "Hello world!"}'
```

```
HTTP/1.1 200 OK
Date: Tue, 08 May 2018 22:16:15 GMT
Content-Length: 27
Content-Type: text/plain; charset=utf-8

{"message": "Hello world!"}
```

Review the Cloud Functions logs:

```
gcloud beta functions logs read echo
```

```
LEVEL  NAME  EXECUTION_ID  TIME_UTC                 LOG
D      echo  vu7o6qv5i8dl  2018-05-08 22:31:20.326  Function execution started
I      echo  vu7o6qv5i8dl  2018-05-08 22:31:20.332  Handling HTTP event b6624b75-ca8a-4c97-86d9-fa98881dfdd8
D      echo  vu7o6qv5i8dl  2018-05-08 22:31:20.337  Function execution took 12 ms, finished with status code: 200
```

## Routing Events to Kubernetes Services

In this section you will deploy the `echo:event-gateway` container using Kubernetes and route cloud events to it using the Event Gateway.

Create a `echo` Kubernetes deployment and service:

```
kubectl create -f deployments/echo.yaml
```

```
deployment "echo" created
service "echo" created
```

Register the Kubernetes `echo` service with the Event Gateway:

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

> At this point the `echo` service has been registered with the Event Gateway, but before it can receive events a subscription must be created.

```
curl -X DELETE \
  http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/subscriptions/http,POST,%2F
```

### Create a Subscription

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

### Emit an event

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4000/ \
  --data '{"message": "Hello World!"}'
```

Review the `echo` container logs:

```
kubectl logs echo-77d48cb484-swxhg
```

```
2018/05/08 22:26:51 Starting HTTP server...
2018/05/08 22:26:57 Handling HTTP event 144685d9-b9da-487c-b5f8-5f13d3f477b8 ...
```
