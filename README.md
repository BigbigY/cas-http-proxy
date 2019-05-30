# cas http proxy

# BUILD
```
GOOS=linux GOARCH=amd64 go build .
```

# RUN
```
./cas-http-proxy -cas-url="https://cas.example.com/cas" -auth-file="./AuthUser.db" -app-addr="IPADDR:5601" -service-addr=":8080"
```