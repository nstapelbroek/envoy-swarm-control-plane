FROM envoyproxy/envoy:v1.28.0

ENV ENVOY_UID=0
EXPOSE 80
EXPOSE 443

COPY config.yml /etc/envoy/envoy.yaml

# Override CMD with a shell so we can expand variables like $HOSTNAME.
CMD ["sh", "-c", "envoy -c /etc/envoy/envoy.yaml --service-node $HOSTNAME"]
