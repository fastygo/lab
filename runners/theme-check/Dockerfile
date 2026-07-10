FROM wordpress:cli-php8.2
USER root
COPY entrypoint.sh /runner/entrypoint.sh
COPY to-findings.php /runner/to-findings.php
COPY run-check.php /runner/run-check.php
RUN chmod +x /runner/entrypoint.sh
WORKDIR /var/www/html
USER www-data
ENTRYPOINT ["/runner/entrypoint.sh"]
