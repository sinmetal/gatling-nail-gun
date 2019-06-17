# gatling-nail-gun
Spannerにおらおらおらおらおらおら更新クエリを投げまくってみるサンプル

## Environment

### Local

```
GCLOUD_PROJECT=
PLAN_QUEUE_NAME=
SPANNER_DATABASE=
GCLOUD_SERVICE_ACCOUNT=hoge@hoge.com
```

## Test

```
curl https://{YOUR RUN URL}/setup/ -X POST \
  -H "Content-Type: application/json\nAuthorization: Bearer $(gcloud config config-helper --format 'value(credential.id_token)')" \
  -d '{"sql": "SELECT Id FROM Tweet WHERE STARTS_WITH(Id, \"%v\") AND Id > \"%v\" ORDER BY Id Limit 1000"}'
```