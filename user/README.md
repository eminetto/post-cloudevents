To test:

```
curl -X "POST" "http://localhost:3000/v1/user" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/json' \
     -d $'{
  "name": "Ozzy Osbourne",
  "password": "12345"
}'

```
