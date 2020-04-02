FROM google/cloud-sdk:latest
Add . /go/src/cen
COPY bin/cen /
ENV PORT 8080
CMD ["./cen"]
