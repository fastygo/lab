FROM wordpress:cli-php8.2
USER root
COPY entrypoint.sh /runner/entrypoint.sh
COPY to-findings.php /runner/to-findings.php
COPY run-check.php /runner/run-check.php
RUN chmod +x /runner/entrypoint.sh \
  && mkdir -p /var/www/html
WORKDIR /var/www/html
# Root: compose volume perms often block www-data from wp-content/upgrade
ENTRYPOINT ["/runner/entrypoint.sh"]
