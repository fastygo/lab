FROM wordpress:cli-php8.2
USER root
COPY entrypoint.sh /runner/entrypoint.sh
COPY seed-unit-test.sh /runner/seed-unit-test.sh
COPY run-check.php /runner/run-check.php
COPY to-findings.php /runner/to-findings.php
RUN chmod +x /runner/entrypoint.sh /runner/seed-unit-test.sh \
  && sed -i 's/\r$//' /runner/entrypoint.sh \
  && sed -i 's/\r$//' /runner/seed-unit-test.sh \
  && sed -i 's/\r$//' /runner/run-check.php \
  && sed -i 's/\r$//' /runner/to-findings.php \
  && mkdir -p /var/www/html
WORKDIR /var/www/html
# Root: compose volume perms often block www-data from wp-content/upgrade
ENTRYPOINT ["/runner/entrypoint.sh"]

