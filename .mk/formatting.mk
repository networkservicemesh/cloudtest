.PHONY: format
format:
	@eval $$(go env); \
	GO111MODULE=on ${GOPATH}/bin/goimports -w -local github.com/networkservicemesh/cloudtest -d `find . -type f -name '*.go' -not -name '*.pb.go' -not -path './vendor/*'`

.PHONY: install-formatter
install-formatter:
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
