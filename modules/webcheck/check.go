package webcheck

import (
	"crypto/tls"
	"net"
	"net/http"
	url2 "net/url"
	"time"
)

func checkURL(client http.Client, url string, warnCodes []int) (result string) {
	pUrl, err := url2.ParseRequestURI(url)
	if err != nil {
		return "fail"
	}

	var addrs []string

	addrs, err = net.LookupHost(pUrl.Host)
	if err != nil || len(addrs) < 1 {
		return "fail"
	}

	var req *http.Request

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return "fail"
	}

	var resp *http.Response

	resp, err = client.Do(req)
	if err != nil || resp == nil {
		return "fail"
	}

	if contains(warnCodes, resp.StatusCode) {
		return "warn"
	}

	if resp.StatusCode >= 400 {
		return "fail"
	}

	return "success"
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
