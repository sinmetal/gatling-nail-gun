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

```
gcloud beta run services update gatling-nail-gun --set-env-vars=FIRE_QUEUE_NAME=projects/{PROJECT_ID}/locations/asia-northeast1/queues/fire,SPANNER_DATABASE=projects/{PROJECT_ID}/instances/{INSTNCE}/databases/{DATABASE}
```

## Test

```
curl https://{YOUR RUN URL}/setup/ -X POST \
  -H "Content-Type: application/json\nAuthorization: Bearer $(gcloud config config-helper --format 'value(credential.id_token)')" \
  -d '{"sql": "SELECT Id FROM Tweet WHERE STARTS_WITH(Id, \"%v\") AND Id > \"%v\" ORDER BY Id Limit 1000"}'
```