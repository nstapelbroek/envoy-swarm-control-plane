FROM envoyproxy/envoy:v1.16-latest

ENV ENVOY_UID=0
EXPOSE 80
EXPOSE 443

COPY config.yml /etc/envoy/envoy.yaml
