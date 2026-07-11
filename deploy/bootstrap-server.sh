#!/usr/bin/env bash
# Bootstrap FastyGo Lab stack on a clean Debian host.
# Usage: ssh cursor-lab 'bash /opt/fastygo/lab/deploy/bootstrap-server.sh'
set -euo pipefail
cd /opt/fastygo/lab
export PATH=/usr/local/go/bin:${PATH:-}

echo "==> Strip CRLF from runner scripts"
find runners -name '*.sh' -o -name '*.php' | while read -r f; do sed -i 's/\r$//' "$f"; done

# Public URL for CSS/JS links in HTML (not the lab CLI baseUrl).
PUBLIC_URL="${LAB_WP_URL:-}"
if [ -z "$PUBLIC_URL" ]; then
  # Prefer first non-loopback IPv4
  IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
  if [ -n "$IP" ]; then
    PUBLIC_URL="http://${IP}:8080"
  else
    PUBLIC_URL="http://127.0.0.1:8080"
  fi
fi
export LAB_WP_URL="$PUBLIC_URL"
echo "==> LAB_WP_URL=${LAB_WP_URL}"

echo "==> Org stack (db + wordpress + wpcli)"
docker compose -f deploy/compose/docker-compose.yml --profile org up -d

echo "==> Quality fixture (nginx :8091)"
docker compose -f deploy/compose/docker-compose.yml --profile quality up -d || true

echo "==> Ensure latte theme active (Apache www-data = uid 33)"
docker compose -f deploy/compose/docker-compose.yml --profile org exec -T -u root wordpress \
  bash -c 'test -f /var/www/html/wp-content/themes/latte/style.css || (unzip -qo /themes/latte.zip -d /tmp && mv /tmp/latte /var/www/html/wp-content/themes/latte); mkdir -p /var/www/html/wp-content/themes/latte/~cache; chown -R 33:33 /var/www/html/wp-content/themes/latte'
docker compose -f deploy/compose/docker-compose.yml --profile org exec -T -u root wpcli \
  wp theme activate latte --allow-root || true

echo "==> Set siteurl/home to public URL"
docker compose -f deploy/compose/docker-compose.yml --profile org exec -T -u root wpcli \
  wp option update siteurl "$LAB_WP_URL" --allow-root
docker compose -f deploy/compose/docker-compose.yml --profile org exec -T -u root wpcli \
  wp option update home "$LAB_WP_URL" --allow-root

echo "==> Status"
docker compose -f deploy/compose/docker-compose.yml --profile org ps
curl -s -o /dev/null -w "WP HTTP %{http_code}\n" http://127.0.0.1:8080/ || true
go run ./apps/cli labs
echo "Done. Public site: ${LAB_WP_URL}"
echo "  cd /opt/fastygo/lab && go run ./apps/cli run -f testdata/manifests/org.lab.yaml"
