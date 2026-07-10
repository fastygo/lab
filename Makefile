.PHONY: test cli demo tidy

test:
	go test ./...

cli:
	go run ./apps/cli $(ARGS)

demo:
	go run ./apps/cli run -f testdata/manifests/demo.lab.yaml

tidy:
	go mod tidy
