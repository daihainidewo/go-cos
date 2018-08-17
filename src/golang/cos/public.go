// Package cos cos
// file create by daihao, time is 2018/8/10 18:19
package cos

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"time"
	"strconv"
	"net/http"
	"net/url"
	"io"
	"strings"
	"sort"
)

const (
	QSignAlgorithm = "sha1"
)

// TXcos
type TXcos struct {
	host      string
	secretID  string
	secretKey string
	appid     string
}

// NewTXcos new TXcos
func NewTXcos(secretID, secretKey, appid, host string) *TXcos {
	ret := &TXcos{
		secretID:  secretID,
		host:      host,
		secretKey: secretKey,
		appid:     appid,
	}
	return ret
}

// Signature Signature
func (c *TXcos) signature(method, url, param, headers, tsign string) string {
	mqkeytime := tsign
	mqsigntime := tsign
	signkey := c.hashHmac([]byte(c.secretKey), []byte(mqkeytime))
	httpstr := strings.ToLower(method) + "\n" + url + "\n" + param + "\n" + headers + "\n"
	str2sign := QSignAlgorithm + "\n" + mqsigntime + "\n" + string(c.sha1Hash(httpstr)) + "\n"
	return string(c.hashHmac(signkey, []byte(str2sign)))
}

// hashHmac hmac-sha1
func (c *TXcos) hashHmac(key, s []byte) []byte {
	h := hmac.New(sha1.New, key)
	h.Write(s)
	return []byte(fmt.Sprintf("%x", h.Sum(nil)))
}

// sha1Hash SHA1-HASH
func (c *TXcos) sha1Hash(s string) []byte {
	h := sha1.New()
	h.Write([]byte(s))
	return []byte(fmt.Sprintf("%x", h.Sum(nil)))
}

// Request Request
func (c *TXcos) request(method, path string, header http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = "https"
	req.URL.Host = c.host
	req.Header = header
	req.Header.Set("Host", c.host)

	qheaderlist := make([]string, 0, len(header))
	for k := range header {
		qheaderlist = append(qheaderlist, strings.ToLower(k))
	}
	sort.Strings(qheaderlist)
	mqHeaderList := strings.Join(qheaderlist, ";")

	httpheader := url.Values{}
	for k := range header {
		httpheader.Set(strings.ToLower(k), header.Get(k))
	}
	httpHeaders := httpheader.Encode()

	query := req.URL.Query()
	qquerylist := make([]string, 0, len(query))
	for k := range query {
		qquerylist = append(qquerylist, strings.ToLower(k))
	}
	sort.Strings(qquerylist)
	mqParamList := strings.Join(qquerylist, ";")
	httpParams := query.Encode()

	tn := time.Now()
	tsign := strconv.FormatInt(tn.Unix(), 10) + ";" + strconv.FormatInt(tn.Add(1 * time.Minute).Unix(), 10)

	mqSign := c.signature(method, path, httpParams, httpHeaders, tsign)
	av := "q-sign-algorithm=" + QSignAlgorithm + "&q-ak=" + c.secretID + "&q-sign-time=" + tsign + "&q-key-time=" + tsign + "&q-header-list=" + mqHeaderList + "&q-url-param-list=" + mqParamList + "&q-signature=" + mqSign
	header.Set("Authorization", av)

	return http.DefaultClient.Do(req)
}

// Sendfile Sendfile
func (c *TXcos) Sendfile(path string, header http.Header, body io.Reader) (*http.Response, error) {
	// TODO
	return c.request("PUT", path, header, body)
}

// Downloadfile Downloadfile
func (c *TXcos) Downloadfile(path string, header http.Header) (*http.Response, error) {
	// TODO
	return c.request("GET", path, header, nil)
}

// GetService GetService
func (c *TXcos) GetService() (*http.Response, error) {
	// TODO 
	return c.request("GET", "/", http.Header{}, nil)
}
