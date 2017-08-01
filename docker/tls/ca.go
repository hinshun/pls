package tls

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

func NewHTTPClient(options tlsconfig.Options) (*http.Client, error) {
	tlsc, err := tlsconfig.Client(options)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsc,
		},
	}, nil
}

func WriteCACertificateToFile(client *http.Client, file *os.File, hostAddress string) error {
	caURL := &url.URL{
		Scheme: "https",
		Host:   hostAddress,
		Path:   "ca",
	}

	resp, err := client.Get(caURL.String())
	if err != nil {
		return stacktrace.Propagate(err, "failed to http get '%s'", caURL.String())
	}
	defer resp.Body.Close()

	caBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stacktrace.Propagate(err, "failed to read certificate response body")
	}

	logrus.Infof("certificate: %s", string(caBytes))
	_, err = file.Write(caBytes)
	if err != nil {
		return stacktrace.Propagate(err, "failed to write certificate to temporary file")
	}

	return nil
}
