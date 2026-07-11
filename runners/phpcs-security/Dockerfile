FROM php:8.2-cli-alpine
USER root
RUN apk add --no-cache unzip curl \
  && curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer \
  && composer global require --no-interaction squizlabs/php_codesniffer:^3.10 pheromone/phpcs-security-audit:^2.0 \
  && /root/.composer/vendor/bin/phpcs --config-set installed_paths /root/.composer/vendor/pheromone/phpcs-security-audit
ENV PATH="/root/.composer/vendor/bin:${PATH}"
COPY entrypoint.sh /runner/entrypoint.sh
RUN sed -i 's/\r$//' /runner/entrypoint.sh && chmod +x /runner/entrypoint.sh
ENTRYPOINT ["/runner/entrypoint.sh"]
