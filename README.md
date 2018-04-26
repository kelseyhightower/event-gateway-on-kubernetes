# Serverless Event Gateway on Kubernetes

## Tutorial

This tutorial assumes you have access to a Kubernetes 1.9.0+ clusters.

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

### Deploy the Event Gateway

```
kubectl apply -f event-gateway.yaml
```

```
kubectl get pods
```
```
NAME              READY     STATUS    RESTARTS   AGE
event-gateway-0   2/2       Running   0          7m
```

```
kubectl get svc
```
```
NAME                  TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)                         AGE
event-gateway         LoadBalancer   10.15.248.210   XX.XXX.XXX.XX   4000:31061/TCP,4001:32247/TCP   11m
```

```
EVENT_GATEWAY_IP=$(kubectl get svc \
  event-gateway \
  -o jsonpath={.status.loadBalancer.ingress[0].ip})
```

### Create a Google Cloud Function

Create the source code:

```
/**
 * Responds to any HTTP request that can provide a "message" field in the body.
 *
 * @param {!Object} req Cloud Function request context.
 * @param {!Object} res Cloud Function response context.
 */
exports.helloWorld = (req, res) => {
  // Example input: {"message": "Hello!"}
  if (req.body.data.message === undefined) {
    // This is an error case, as "message" is required.
    res.status(400).send('No message defined!');
  } else {
    // Everything is okay.
    console.log(req.body.data.message);
    res.status(200).send('Success: ' + req.body.data.message);
  }
};
```

### Register the Helloworld Goole Cloud Function

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "helloworld",
    "type": "http",
    "provider":{
      "type": "http",
      "url": "https://us-central1-hightowerlabs.cloudfunctions.net/helloworld"
    }
}'
```

### Create a subscription

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4001/v1/spaces/default/subscriptions \
  --header 'content-type: application/json' \
  --data '{
    "functionId": "helloworld",
    "event": "user.created",
    "path": "/"
  }'
```

### Emit an event

```
curl --request POST \
  --url http://${EVENT_GATEWAY_IP}:4000/ \
  --header 'content-type: application/json' \
  --header 'event: user.created' \
  --data '{"message": "Hello!"}'
```
