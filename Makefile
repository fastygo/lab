.PHONY: test cli demo quality org sec tidy runners org-up quality-up

test:
	go test ./...

cli:
	go run ./apps/cli $(ARGS)

demo:
	go run ./apps/cli run -f testdata/manifests/demo.lab.yaml

quality:
	go run ./apps/cli run -f testdata/manifests/quality.lab.yaml

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

org-up:
	docker compose -f deploy/compose/docker-compose.yml --profile org up -d

quality-up:
	docker compose -f deploy/compose/docker-compose.yml --profile quality up -d
