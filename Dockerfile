ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc
ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/eseries_exporter /eseries_exporter
EXPOSE 9313
ENTRYPOINT ["/eseries_exporter"]
