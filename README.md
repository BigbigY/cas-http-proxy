# cas http proxy

# BUILD
```
GOOS=linux GOARCH=amd64 go build .
```

# RUN
```
./cas-http-proxy -cas-url="https://cas.example.com/cas" -auth-file="./AuthUser.db" -app-addr="IPADDR:5601" -service-addr=":8080"
```

# NGINX PROXY TO CAS PROXY
```
proxy_set_header Host "cas-proxy.example.com";
```