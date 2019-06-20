# gatling-nail-gun
Spannerにおらおらおらおらおらおら更新クエリを投げまくってみるサンプル

## Environment

### Local

```
GCLOUD_PROJECT=
FIRE_QUEUE_NAME=
SPANNER_DATABASE=
GCLOUD_SERVICE_ACCOUNT=hoge@hoge.com
```

## Deploy

### Cloud Tasks

```
gcloud tasks queues create fire \
    --max-doublings=1 --min-backoff=1s \
    --max-backoff=3600s \
    --max-dispatches-per-second=500 \
    --max-concurrent-dispatches=500
```

### Cloud Run

```
gcloud beta run services update gatling-nail-gun --set-env-vars=FIRE_QUEUE_NAME=projects/{PROJECT_ID}/locations/asia-northeast1/queues/fire,SPANNER_DATABASE=projects/{PROJECT_ID}/instances/{INSTNCE}/databases/{DATABASE}
```

## Execute

```
curl https://{YOUR RUN URL}/setup/ -X POST \
  -H "Content-Type: application/json\nAuthorization: Bearer $(gcloud config config-helper --format 'value(credential.id_token)')" \
  -d '{"sql": "SELECT Id FROM Tweet WHERE Id >= \"%v\" ORDER BY Id Limit %v", "schemaVersion": 1, "limit": 1000}'
```