package webcheck

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	url2 "net/url"
	"time"

	"github.com/wtfutil/wtf/logger"
)

const (
	msgSuccess = "success"
	msgWarn    = "warn"
	msgFail    = "fail"
)

func checkURL(client http.Client, url string, warnCodes []int) (result string) {
	pUrl, err := url2.ParseRequestURI(url)
	if err != nil {
		return msgFail
	}

	var addrs []string

	addrs, err = net.LookupHost(pUrl.Host)
	if err != nil || len(addrs) < 1 {
		logger.Log(fmt.Sprintf("%s | failed to resolve: %s", moduleName, pUrl.Host))
		return msgFail
	}

	var req *http.Request

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Log(fmt.Sprintf("%s | failed to create request for: %s", moduleName, url))
		return msgFail
	}

	var resp *http.Response

	resp, err = client.Do(req)
	if err != nil {
		logger.Log(fmt.Sprintf("%s | error returned for: %s error: %+v", moduleName, url, err.Error()))

		return msgFail
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp == nil {
		logger.Log(fmt.Sprintf("%s | empty response for: %s", moduleName, url))

		return msgFail
	}

	if contains(warnCodes, resp.StatusCode) {
		logger.Log(fmt.Sprintf("%s | warn for: %s due to status code: %d", moduleName, url, resp.StatusCode))
		return msgWarn
	}

	if resp.StatusCode >= 400 {
		logger.Log(fmt.Sprintf("%s | fail for: %s due to status code: %d", moduleName, url, resp.StatusCode))
		return msgFail
	}

	return msgSuccess
}

func getClient(settings *Settings) http.Client {
	tr := &http.Transport{
		TLSHandshakeTimeout:   5 * time.Second,
		DisableKeepAlives:     true,
		DisableCompression:    true,
		ResponseHeaderTimeout: 3 * time.Second,
	}

	if settings.ignoreBadSSL {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	cr := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	if settings.followRedirects {
		cr = nil
	}

	return http.Client{
		Transport:     tr,
		Timeout:       5 * time.Second,
		CheckRedirect: cr,
	}
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
