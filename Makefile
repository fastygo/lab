.PHONY: test cli demo quality quality-wp static-web org sec tidy runners org-up quality-up org-seed

test:
	go test ./...

cli:
	go run ./apps/cli $(ARGS)

demo:
	go run ./apps/cli run -f testdata/manifests/demo.lab.yaml

quality:
	go run ./apps/cli run -f testdata/manifests/quality.lab.yaml

quality-wp:
	go run ./apps/cli run -f testdata/manifests/quality-wp.lab.yaml

static-web:
	go run ./apps/cli run -f testdata/manifests/quality-staticweb.lab.yaml

org:
	go run ./apps/cli run -f testdata/manifests/org.lab.yaml

sec:
	go run ./apps/cli run -f testdata/manifests/sec.lab.yaml

tidy:
	go mod tidy

runners:
	docker build -t lab/lighthouse:local runners/lighthouse
	docker build -t lab/axe:local runners/axe
	docker build -t lab/theme-check:local runners/theme-check
	docker build -t lab/vnu:local runners/vnu
	docker build -t lab/wpscan:local runners/wpscan
	docker build -t lab/composer-audit:local runners/composer-audit
	docker build -t lab/nuclei:local runners/nuclei
	docker build -t lab/notice-hunter:local runners/notice-hunter
	docker build -t lab/org-keyboard:local runners/org-keyboard
	docker build -t lab/css-lint:local runners/css-lint
	docker build -t lab/quality-extras:local runners/quality-extras

org-up:
	docker compose -f deploy/compose/docker-compose.yml --profile org up -d

quality-up:
	docker compose -f deploy/compose/docker-compose.yml --profile quality up -d

# Import Theme Unit Test XML into compose WP and write testdata/fixtures/org-seed.json
org-seed:
	docker compose -f deploy/compose/docker-compose.yml --profile org exec -T -u root wpcli \
		sh -c 'wp plugin install wordpress-importer --activate --force --allow-root >/dev/null 2>&1 || true; \
		wp option get lab_unit_test_imported --allow-root >/dev/null 2>&1 || \
		(wp import /lab/fixtures/themeunittestdata.wordpress.xml --authors=create --allow-root && \
		 wp option update lab_unit_test_imported 1 --allow-root); \
		wp rewrite structure "/%postname%/" --allow-root; wp rewrite flush --hard --allow-root; \
		ATTACH=$$(wp post list --post_type=attachment --post_mime_type=image --fields=ID --format=ids --allow-root | awk "{print \$$1}"); \
		POST=$$(wp post list --post_type=post --post_status=publish --fields=ID --format=ids --allow-root | awk "{print \$$1}"); \
		PAGE=$$(wp post list --post_type=page --post_status=publish --fields=ID --format=ids --allow-root | awk "{print \$$1}"); \
		TAG=$$(wp term list post_tag --field=slug --allow-root | head -1); \
		CAT=$$(wp term list category --field=term_id --allow-root | head -1); \
		printf "%s\n" "{\"attachmentId\":\"$$ATTACH\",\"postId\":\"$${POST:-1}\",\"pageId\":\"$${PAGE:-2}\",\"catId\":\"$${CAT:-1}\",\"tagSlug\":\"$${TAG:-test}\",\"imported\":true}" \
		  > /var/www/html/wp-content/lab-org-seed.json; \
		cat /var/www/html/wp-content/lab-org-seed.json'
	docker compose -f deploy/compose/docker-compose.yml --profile org cp \
		wpcli:/var/www/html/wp-content/lab-org-seed.json testdata/fixtures/org-seed.json
