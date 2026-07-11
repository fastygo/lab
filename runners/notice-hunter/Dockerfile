FROM wordpress:cli-php8.2
USER root
RUN apk add --no-cache curl 2>/dev/null || (apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq curl >/dev/null && rm -rf /var/lib/apt/lists/*)
COPY entrypoint.sh /runner/entrypoint.sh
RUN chmod +x /runner/entrypoint.sh && sed -i 's/\r$//' /runner/entrypoint.sh
WORKDIR /var/www/html
ENTRYPOINT ["/runner/entrypoint.sh"]
