FROM google/cloud-sdk:latest
Add . /go/src/contact-tracing
COPY bin/contact-tracing /
COPY certs/www.wolk.com.key /tmp/
COPY certs/www.wolk.com.bundle /tmp/
COPY conf/ct.conf /tmp/
ENV PORT 8080
ENV SSLDIR /tmp
ENV CTDIR /tmp
CMD ["./contact-tracing"]
