// Package cos cos
// file create by daihao, time is 2018/8/10 18:19
package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
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
func NewTXcos(conf *TXConf) *TXcos {
	ret := &TXcos{
		secretID:  conf.SecretId,
		host:      conf.Host,
		secretKey: conf.SecretKey,
		appid:     conf.AppId,
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
func (c *TXcos) NewRequest(method, path string, value url.Values, header http.Header, body io.Reader) (*TXcosRequest) {
	ret := &TXcosRequest{
		c:      c,
		header: header,
		value:  value,
		method: method,
		path:   path,
		body:   body,
		client: http.DefaultClient,
	}
	ret.header.Add("Host", c.host)
	return ret
}

// Delete Delete
func (t *TXcos) Delete(path string) (error) {
	resp, err := t.NewRequest("DELETE", path, url.Values{}, http.Header{}, nil).Do()
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(data))
	}
	return nil
}

func (t *TXcos) Get(path string) ([]byte, error) {
	resp, err := t.NewRequest("GET", path, url.Values{}, http.Header{}, nil).Do()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, errors.New(string(data))
	}
	return data, nil
}

func (t *TXcosRequest) Head(path string) (http.Header, error) {
	resp, err := t.c.NewRequest("HEAD", path, url.Values{}, http.Header{}, nil).Do()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(data))
	}
	return resp.Header, nil
}

func (t *TXcos) Put(path string, body io.Reader, contentLength int) (error) {
	header := http.Header{}
	header.Set("Content-Length", strconv.Itoa(contentLength))
	resp, err := t.NewRequest("PUT", path, url.Values{}, header, body).Do()
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(data))
	}
	return nil
}

// TXcosRequest
type TXcosRequest struct {
	c      *TXcos
	header http.Header
	value  url.Values
	method string
	path   string
	body   io.Reader
	client *http.Client
}

// Do Do
func (t *TXcosRequest) Do() (*http.Response, error) {
	path := t.path
	if len(t.value) != 0 {
		path += "?" + t.value.Encode()
	}
	req, err := http.NewRequest(t.method, path, t.body)
	if err != nil {
		return nil, err
	}
	req.URL.Scheme = "https"
	req.URL.Host = t.c.host
	req.Header = t.header
	req.Header.Set("Host", t.c.host)

	qheaderlist := make([]string, 0, len(t.header))
	for k := range t.header {
		qheaderlist = append(qheaderlist, strings.ToLower(k))
	}
	sort.Strings(qheaderlist)
	mqHeaderList := strings.Join(qheaderlist, ";")

	httpheader := url.Values{}
	for k := range t.header {
		httpheader.Set(strings.ToLower(k), t.header.Get(k))
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

	mqSign := t.c.signature(t.method, t.path, httpParams, httpHeaders, tsign)
	av := "q-sign-algorithm=" + QSignAlgorithm + "&q-ak=" + t.c.secretID + "&q-sign-time=" + tsign + "&q-key-time=" + tsign + "&q-header-list=" + mqHeaderList + "&q-url-param-list=" + mqParamList + "&q-signature=" + mqSign
	req.Header.Set("Authorization", av)

	return t.client.Do(req)
}

func (t *TXcosRequest) SetValueContentType(s string) *TXcosRequest {
	t.value.Add("response-content-type", s)
	return t
}

func (t *TXcosRequest) SetValueContentLanguage(s string) *TXcosRequest {
	t.value.Add("response-content-language", s)
	return t
}
func (t *TXcosRequest) SetValueEexpires(s string) *TXcosRequest {
	t.value.Add("response-expires", s)
	return t
}
func (t *TXcosRequest) SetValueCacheControl(s string) *TXcosRequest {
	t.value.Add("response-cache-control", s)
	return t
}
func (t *TXcosRequest) SetValueContentEncoding(s string) *TXcosRequest {
	t.value.Add("response-content-encoding", s)
	return t
}

// SetHeaderAuthorization SetHeaderAuthorization
func (t *TXcosRequest) SetHeaderAuthorization(s string) (*TXcosRequest) {
	t.header.Set("Authorization", s)
	return t
}
func (t *TXcosRequest) SetHeaderContentLength(s string) (*TXcosRequest) {
	t.header.Set("Content-Length", s)
	return t
}
func (t *TXcosRequest) SetHeaderContentType(s string) (*TXcosRequest) {
	t.header.Set("Content-Type", s)
	return t
}
func (t *TXcosRequest) SetHeaderContentMD5(s string) (*TXcosRequest) {
	t.header.Set("Content-MD5", s)
	return t
}
func (t *TXcosRequest) SetHeaderDate(s string) (*TXcosRequest) {
	t.header.Set("Date", s)
	return t
}
func (t *TXcosRequest) SetHeaderExpect(s string) (*TXcosRequest) {
	t.header.Set("Expect", s)
	return t
}

// SetHeaderRange SetHeaderRange
func (t *TXcosRequest) SetHeaderRange(s string) (*TXcosRequest) {
	t.header.Add("Range", s)
	return t
}

// SetHeaderIfUnmodifiedSince SetHeaderIfUnmodifiedSince
func (t *TXcosRequest) SetHeaderIfUnmodifiedSince(s string) (*TXcosRequest) {
	t.header.Add("If-Unmodified-Since", s)
	return t
}

// SetHeaderIfMatch SetHeaderIfMatch
func (t *TXcosRequest) SetHeaderIfMatch(s string) (*TXcosRequest) {
	t.header.Add("If-Match", s)
	return t
}

// SetHeaderIfNoneMatch SetHeaderIfNoneMatch
func (t *TXcosRequest) SetHeaderIfNoneMatch(s string) (*TXcosRequest) {
	t.header.Add("If-None-Match", s)
	return t
}

// SetHeaderIfModifiedSince SetHeaderIfModifiedSince
func (t *TXcosRequest) SetHeaderIfModifiedSince(s string) (*TXcosRequest) {
	t.header.Set("If-Modified-Since", s)
	return t
}

// SetHeaderContentDisposition SetHeaderContentDisposition
func (t *TXcosRequest) SetHeaderContentDisposition(s string) (*TXcosRequest) {
	t.header.Set("Content-Disposition", s)
	return t
}

func (t *TXcosRequest) SetHeaderContentEncoding(s string) (*TXcosRequest) {
	t.header.Set("Content-Encoding", s)
	return t
}

func (t *TXcosRequest) SetHeaderExpires(s string) (*TXcosRequest) {
	t.header.Set("Expires", s)
	return t
}
func (t *TXcosRequest) SetHeaderMeta(key, value string) (*TXcosRequest) {
	t.header.Set("x-cos-meta-"+key, value)
	return t
}
func (t *TXcosRequest) SetHeaderstorageclass(s string) (*TXcosRequest) {
	t.header.Set("x-cos-storage-class", s)
	return t
}
func (t *TXcosRequest) SetHeaderAcl(s string) (*TXcosRequest) {
	t.header.Set("x-cos-acl", s)
	return t
}
func (t *TXcosRequest) SetHeaderGrantRead(s string) (*TXcosRequest) {
	t.header.Set("x-cos-grant-read", s)
	return t
}
func (t *TXcosRequest) SetHeaderGrantWrite(s string) (*TXcosRequest) {
	t.header.Set("x-cos-grant-write", s)
	return t
}
func (t *TXcosRequest) SetHeaderGrantFullControl(s string) (*TXcosRequest) {
	t.header.Set("x-cos-grant-full-control", s)
	return t
}
func (t *TXcosRequest) SetHeaderServerSideEncryption(s string) (*TXcosRequest) {
	t.header.Set("x-cos-server-side-encryption", s)
	return t
}
