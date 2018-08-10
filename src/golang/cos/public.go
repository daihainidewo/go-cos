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
	"io/ioutil"
)

const (
	QSignAlgorithm = "q-sign-algorithm"
	QAk            = "q-ak"
	QSignTime      = "q-sign-time"
	QKeyTime       = "q-key-time"
	QHeaderList    = "q-header-list"
	QUrlParamList  = "q-url-param-list"
	QSignature     = "q-signature"
)

// TXcos
type TXcos struct {
	host      string
	secretID  string
	secretKey string
	appid     string
	KVmap     map[string]string
}

// NewTXcos new TXcos
func NewTXcos(secretID, secretKey, appid, host string) *TXcos {
	ret := &TXcos{
		KVmap:     make(map[string]string),
		secretID:  secretID,
		host:      host,
		secretKey: secretKey,
		appid:     appid,
	}
	ret.init()
	return ret
}

// init init
func (c *TXcos) init() {
	c.KVmap[QSignAlgorithm] = "sha1"
	c.KVmap[QAk] = c.secretID
	c.KVmap[QSignTime] = ""
	c.KVmap[QKeyTime] = ""
	c.KVmap[QHeaderList] = ""
	c.KVmap[QUrlParamList] = ""
	c.KVmap[QSignature] = ""
}

// Signature Signature
func (c *TXcos) signature(method, uri, param, headers string) {
	tn := time.Now()
	c.KVmap[QSignTime] = strconv.FormatInt(tn.Unix(), 10) + ";" + strconv.FormatInt(tn.Add(1 * time.Minute).Unix(), 10)
	c.KVmap[QKeyTime] = c.KVmap[QSignTime]
	signkey := c.hashHmac([]byte(c.KVmap[QKeyTime]), []byte(c.secretKey))
	httpstr := fmt.Sprintf("%s\n%s\n%s\n%s\n", method, uri, param, headers)
	str2sign := fmt.Sprintf("%s\n%s\n%s\n", c.KVmap[QSignAlgorithm], c.KVmap[QSignTime], string(c.sha1Hash(httpstr)))
	signature := c.hashHmac([]byte(str2sign), signkey)
	c.KVmap[QSignature] = string(signature)
}

// hashHmac hmac-sha1
func (c *TXcos) hashHmac(s, key []byte) []byte {
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
	//req, err := http.NewRequest(method, path, body)
	//if err != nil {
	//	return nil, err
	//}
	//req.Header = header
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	req := &http.Request{
		Method: method,
		URL: &url.URL{
			Scheme: "https",
			Path:   path,
			Host:   c.host,
		},
		Header: header,
		Body:   rc,
	}
	req.ContentLength = 11
	return http.DefaultClient.Do(req)
}

// Sendfile Sendfile
func (c *TXcos) Sendfile(path string, header http.Header, body io.Reader) (*http.Response, error) {
	// TODO
	hlist := make([]string, 0, len(header))
	for k := range header {
		hlist = append(hlist, strings.ToLower(k))
	}
	sort.Strings(hlist)
	c.KVmap[QHeaderList] = strings.Join(hlist, ";")
	hheader := url.Values{}
	for k := range header {
		hheader.Set(strings.ToLower(k), header.Get(k))
	}
	c.signature("PUT", path, "", hheader.Encode())
	av := fmt.Sprintf("q-sign-algorithm=%s&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=%s&q-url-param-list=%s&q-signature=%s", c.KVmap[QSignAlgorithm], c.KVmap[QAk], c.KVmap[QSignTime], c.KVmap[QKeyTime], c.KVmap[QHeaderList], c.KVmap[QUrlParamList], c.KVmap[QSignature])
	header.Set("Authorization", av)
	return c.request("put", path, header, body)
}

// Downloadfile Downloadfile
func (c *TXcos) Downloadfile(path string, header http.Header) (*http.Response, error) {
	// TODO
	return c.request("get", path, header, nil)
}
