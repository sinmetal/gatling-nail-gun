# gatling-nail-gun
Spannerにおらおらおらおらおらおら更新クエリを投げまくってみるサンプル

## Test

```
curl https://{YOUR RUN URL}/setup/ -X POST \
  -H "Content-Type: application/json\nAuthorization: Bearer $(gcloud config config-helper --format 'value(credential.id_token)')" \
  -d '{"sql": "SELECT Id FROM Tweet WHERE STARTS_WITH(Id, '%v')"}'
```