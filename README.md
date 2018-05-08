# Serverless Event Gateway on Kubernetes

This guide will walk you through provisioning a multi-node [Event Gateway](https://github.com/serverless/event-gateway) cluster on Kubernetes. The goal of this guide is to introduce you to the Event Gateway and get a feel for the role it plays in a Serverless Architecture.

## Tutorial

This tutorial assumes you have access to a Kubernetes 1.9.6+ cluster and [Google Cloud Functions](https://cloud.google.com/functions)*.

```
gcloud container clusters create event-gateway \
  --async \
  --enable-autorepair \
  --enable-network-policy \
  --cluster-version 1.9.6-gke.1 \
  --machine-type n1-standard-2 \
  --num-nodes 3 \
  --zone us-west1-c
```

> Any backend that can respond to HTTP request will work. Google Cloud Functions is only being used to streamline the learning process.

### Deploy the Event Gateway

In this section you will create a three node Event Gateway cluster backed by a single node etcd cluster.

Create the `etcd` statefulset:

```
kubectl apply -f statefulsets/etcd.yaml
```

Create the `event-gateway` deployment:

```
kubectl apply -f deployments/event-gateway.yaml
```

At this point the Event Gateway should be up and running and exposed via an external loadbalancer.

```
kubectl get pods
```
```
NAME                            READY     STATUS    RESTARTS   AGE
etcd-0                          1/1       Running   0          1m
event-gateway-cff6df9cd-89dlr   1/1       Running   0          30s
event-gateway-cff6df9cd-f42q6   1/1       Running   0          30s
event-gateway-cff6df9cd-mrtfs   1/1       Running   0          30s
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

### Create a Google Cloud Function

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

### Register the Helloworld Goole Cloud Function

In this section you will register the `helloworld` function with the Event Gateway.

Register the `helloworld` function:

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

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data @register-function.json
```

### Create a subscription

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

### Emit an event

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
