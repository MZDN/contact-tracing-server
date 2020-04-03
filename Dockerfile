FROM google/cloud-sdk:latest
Add . /go/src/cen
COPY bin/cen /
COPY certs/www.wolk.com.key /tmp/
COPY certs/www.wolk.com.bundle /tmp/
COPY conf/cen.conf /tmp/
ENV PORT 8080
ENV SSLDIR /tmp
ENV TMPDIR /tmp
CMD ["./cen"]
