FROM wordpress:cli-php8.2
COPY entrypoint.sh /runner/entrypoint.sh
USER root
RUN chmod +x /runner/entrypoint.sh
USER www-data
ENTRYPOINT ["/runner/entrypoint.sh"]
