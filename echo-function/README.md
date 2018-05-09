# echo

The `echo` [cloud function](https://cloud.google.com/functions) handles [HTTP Events](https://github.com/serverless/event-gateway/blob/master/docs/api.md#http-event) based on the [CloudEvent](https://github.com/serverless/event-gateway/blob/master/docs/api.md#event-definition) event format and echoes the request body back to the caller.

## Usage

```
gcloud beta functions deploy echo \
  --source echo-function \
  --trigger-http
```

```
gcloud beta functions logs read echo
```

```
LEVEL  NAME  EXECUTION_ID  TIME_UTC                 LOG
D      echo  4uczimni6d70  2018-05-08 23:24:11.206  Function execution started
I      echo  4uczimni6d70  2018-05-08 23:24:11.458  Handling HTTP event 13e2cfa2-3c86-42dc-a8be-a01648b6444c
D      echo  4uczimni6d70  2018-05-08 23:24:11.546  Function execution took 341 ms, finished with status code: 200
```

```
2018/05/08 23:30:04 Starting HTTP server...
2018/05/08 23:35:35 Handling HTTP event f3a37c57-d85a-4942-b92c-cef56713d538 ...
```
