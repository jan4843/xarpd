release: \
	linux/amd64 \
	linux/arm \
	linux/arm64 \

clean:
	$(RM) -r release

NAME := $(shell go list)
VERSION := $(shell git name-rev --tags --name-only HEAD)
DISTS := $(shell go tool dist list)
$(DISTS): GOOS = $(firstword $(subst /, ,$@))
$(DISTS): GOARCH = $(lastword $(subst /, ,$@))
$(DISTS):
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags="-buildid= -s -w" -trimpath -o release/$(NAME)-$(VERSION)-$(GOOS)-$(GOARCH)
