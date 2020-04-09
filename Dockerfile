FROM google/cloud-sdk:latest
Add . /go/src/findmypk
COPY bin/findmypk /
COPY certs/www.wolk.com.key /tmp/
COPY certs/www.wolk.com.bundle /tmp/
COPY conf/fmpk.conf /tmp/
ENV PORT 8080
ENV SSLDIR /tmp
ENV FMPKDIR /tmp
CMD ["./findmypk"]
