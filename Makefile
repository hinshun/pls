pls: vendor
	go build pls.go

vendor:
	glide install --strip-vendor
	gofmt -w -r '"github.com/Sirupsen/logrus" -> "github.com/sirupsen/logrus"' ./

.PHONY: pls

