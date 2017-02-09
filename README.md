Generate fake certificate
```
openssl req -new -x509 -days 3650 -extensions v3_ca -keyout cakey.pem -out cacert.pem
openssl req -new -nodes -out my_req.pem -keyout my.key
openssl x509 -req -days 365 -in my_req.pem -CA cacert.pem -CAkey cakey.pem -set_serial 01 -out my.pem
```

Use
```
./mitm-proxy -crt=my.pem -key=my.key -local_addr=0.0.0.0:15653 -remote_addr=example.com:15653
```
