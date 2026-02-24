OUT_DIR := ./out
VERSION := "1.0.0"

# This requires to install the musl cross compiler brew install FiloSottile/musl-cross/musl-cross
# when cross compiling linux on darwin
build-release-on-darwin:
	CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build --ldflags '-linkmode external -extldflags=-static' -tags musl -o out/chain_sink_$(VERSION)_linux_amd64 ./cmd/chain_sink
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o out/chain_sink_$(VERSION)_darwin_arm64 ./cmd/chain_sink
	sha1 out/chain_sink_$(VERSION)_linux_amd64 out/chain_sink_$(VERSION)_darwin_arm64 > out/chain_sink_$(VERSION).sha1
	zip -j out/chain_sink_$(VERSION).zip out/chain_sink_$(VERSION)_linux_amd64 out/chain_sink_$(VERSION)_darwin_arm64 out/chain_sink_$(VERSION).sha1
