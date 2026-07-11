FROM composer:2
USER root
RUN apk add --no-cache unzip 2>/dev/null || (apt-get update && apt-get install -y --no-install-recommends unzip && rm -rf /var/lib/apt/lists/*) || true
COPY entrypoint.sh /runner/entrypoint.sh
RUN sed -i 's/\r$//' /runner/entrypoint.sh && chmod +x /runner/entrypoint.sh
ENTRYPOINT ["/runner/entrypoint.sh"]
