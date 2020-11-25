FROM netapp/eseries-webservices:4.20 as base
ADD wsconfig.xml ./wsconfig.xml
LABEL autodelete="true"
FROM alpine:3.10
WORKDIR /opt/netapp/webservices_proxy
RUN apk update && apk upgrade && rm -rf /var/cache/apk/*
RUN apk add openjdk8-jre
COPY --from=base /opt/netapp/webservices_proxy ./
CMD ["./webserver.sh"]