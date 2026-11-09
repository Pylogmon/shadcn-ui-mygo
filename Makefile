export GOCACHE := $(CURDIR)/.cache/go-build

.PHONY: test vet run build app popups-app snapshots generate verify-consumer
test:
	go test ./...
vet:
	go vet ./...
run:
	go run ./cmd/gallery
build:
	mkdir -p build
	CGO_ENABLED=0 go build -o build/gallery ./cmd/gallery
app:
	sh scripts/package-macos.sh
popups-app:
	sh scripts/package-macos.sh native-popups
snapshots:
	mkdir -p artifacts
	go run ./cmd/gallery -snapshot artifacts/gallery-light.png
	go run ./cmd/gallery -dark -snapshot artifacts/gallery-dark.png
	go run ./cmd/gallery -page themes -snapshot artifacts/gallery-themes.png
generate:
	go generate ./...
verify-consumer:
	go run ./cmd/verify-consumer
