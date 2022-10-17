FROM envoyproxy/envoy:v1.23-latest

ENV ENVOY_UID=0
EXPOSE 80
EXPOSE 443

COPY config.yml /etc/envoy/envoy.yaml

# Override CMD with a shell so we can expand variables like $HOSTNAME.
# I did this to allow scaling the image without adding more and more layers of complexity to keep tracing etc.
CMD ["sh", "-c", "envoy -c /etc/envoy/envoy.yaml --service-node $HOSTNAME"]
