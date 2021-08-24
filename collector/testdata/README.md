Generate fake CA certs:

```
openssl genrsa -out rootCA.key 4096
openssl req -x509 -new -key rootCA.key -days 3650 -out rootCA.crt
```
