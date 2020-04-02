FROM google/cloud-sdk:latest
COPY . /cen
RUN make /cen

ENV PORT 8080
CMD ["/cen/bin/cen"]
