// Package httpbin providers HTTP handlers for httpbin.org endpoints and a
// multiplexer to directly hook it up to any http.Server or httptest.Server.
package httpbin

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

var (
	// BinaryChunkSize is buffer length used for stuff like generating
	// large blobs.
	BinaryChunkSize = 64 * 1024

	// DelayMax is the maximum execution time for /delay endpoint.
	DelayMax = 10 * time.Second

	// StreamInterval is the default interval between writing objects to the stream.
	StreamInterval = 1 * time.Second
)

// GetMux returns the mux with handlers for httpbin endpoints registered.
func GetMux() *mux.Router {

	r := mux.NewRouter()
	r.HandleFunc(`/`, HomeHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/ip`, IPHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/user-agent`, UserAgentHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/headers`, HeadersHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/get`, GetHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/post`, PostHandler).Methods(http.MethodPost)
	r.HandleFunc(`/redirect/{n:[\d]+}`, RedirectHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/absolute-redirect/{n:[\d]+}`, AbsoluteRedirectHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/redirect-to`, RedirectToHandler).Methods(http.MethodGet, http.MethodHead).Queries("url", "{url:.+}")
	r.HandleFunc(`/status/{code:[\d]+}`, StatusHandler)
	r.HandleFunc(`/bytes/{n:[\d]+}`, BytesHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/delay/{n:\d+(?:\.\d+)?}`, DelayHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/stream/{n:[\d]+}`, StreamHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/drip`, DripHandler).Methods(http.MethodGet, http.MethodHead).Queries(
		"numbytes", `{numbytes:\d+}`,
		"duration", `{duration:\d+(?:\.\d+)?}`)
	r.HandleFunc(`/cookies`, CookiesHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/cookies/set`, SetCookiesHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/cookies/delete`, DeleteCookiesHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/cache`, CacheHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/cache/{n:[\d]+}`, SetCacheHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/gzip`, GZIPHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/deflate`, DeflateHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/html`, HTMLHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/xml`, XMLHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/robots.txt`, RobotsTXTHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/deny`, DenyHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/basic-auth/{u}/{p}`, BasicAuthHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/hidden-basic-auth/{u}/{p}`, HiddenBasicAuthHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/image/gif`, GIFHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/image/png`, PNGHandler).Methods(http.MethodGet, http.MethodHead)
	r.HandleFunc(`/image/jpeg`, JPEGHandler).Methods(http.MethodGet, http.MethodHead)
	return r
}

// HomeHandler serves static HTML content for the index page.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta http-equiv='content-type' value='text/html;charset=utf8'>
  <meta name='generator' value='Ronn/v0.7.3 (http://github.com/rtomayko/ronn/tree/0.7.3)'>
  <title>go-httpbin(1): HTTP Client Testing Service</title>
  <style type='text/css' media='all'>
  /* style: man */
  body#manpage {margin:0}
  .mp {max-width:100ex;padding:0 9ex 1ex 4ex}
  .mp p,.mp pre,.mp ul,.mp ol,.mp dl {margin:0 0 20px 0}
  .mp h2 {margin:10px 0 0 0}
  .mp > p,.mp > pre,.mp > ul,.mp > ol,.mp > dl {margin-left:8ex}
  .mp h3 {margin:0 0 0 4ex}
  .mp dt {margin:0;clear:left}
  .mp dt.flush {float:left;width:8ex}
  .mp dd {margin:0 0 0 9ex}
  .mp h1,.mp h2,.mp h3,.mp h4 {clear:left}
  .mp pre {margin-bottom:20px}
  .mp pre+h2,.mp pre+h3 {margin-top:22px}
  .mp h2+pre,.mp h3+pre {margin-top:5px}
  .mp img {display:block;margin:auto}
  .mp h1.man-title {display:none}
  .mp,.mp code,.mp pre,.mp tt,.mp kbd,.mp samp,.mp h3,.mp h4 {font-family:monospace;font-size:14px;line-height:1.42857142857143}
  .mp h2 {font-size:16px;line-height:1.25}
  .mp h1 {font-size:20px;line-height:2}
  .mp {text-align:justify;background:#fff}
  .mp,.mp code,.mp pre,.mp pre code,.mp tt,.mp kbd,.mp samp {color:#131211}
  .mp h1,.mp h2,.mp h3,.mp h4 {color:#030201}
  .mp u {text-decoration:underline}
  .mp code,.mp strong,.mp b {font-weight:bold;color:#131211}
  .mp em,.mp var {font-style:italic;color:#232221;text-decoration:none}
  .mp a,.mp a:link,.mp a:hover,.mp a code,.mp a pre,.mp a tt,.mp a kbd,.mp a samp {color:#0000ff}
  .mp b.man-ref {font-weight:normal;color:#434241}
  .mp pre {padding:0 4ex}
  .mp pre code {font-weight:normal;color:#434241}
  .mp h2+pre,h3+pre {padding-left:0}
  ol.man-decor,ol.man-decor li {margin:3px 0 10px 0;padding:0;float:left;width:33%;list-style-type:none;text-transform:uppercase;color:#999;letter-spacing:1px}
  ol.man-decor {width:100%}
  ol.man-decor li.tl {text-align:left}
  ol.man-decor li.tc {text-align:center;letter-spacing:4px}
  ol.man-decor li.tr {text-align:right;float:right}
  </style>
  <style type='text/css' media='all'>
  /* style: 80c */
  .mp {max-width:86ex}
  ul {list-style: None; margin-left: 1em!important}
  .man-navigation {left:101ex}
  </style>
</head>

<body id='manpage'>


<div class='mp'>
<h1>go-httpbin(1)</h1>
<p>A golang port of the venerable <a href="https://httpbin.org/">httpbin.org</a> HTTP request &amp; response testing service.</p>

<h2 id="ENDPOINTS">ENDPOINTS</h2>

<ul>
<li><a href="/" data-bare-link="true"><code>/</code></a> This page.</li>
<li><a href="ip" data-bare-link="true"><code>/ip</code></a> Returns Origin IP.</li>
<li><a href="user-agent" data-bare-link="true"><code>/user-agent</code></a> Returns user-agent.</li>
<li><a href="headers" data-bare-link="true"><code>/headers</code></a> Returns header dict.</li>
<li><a href="get" data-bare-link="true"><code>/get</code></a> Returns GET data.</li>
<li><code>post</code> Returns POST data.</li>
<li><code>patch</code> Returns PATCH data.</li>
<li><code>put</code> Returns PUT data.</li>
<li><code>delete</code> Returns DELETE data</li>
<li><a href="encoding/utf8"><code>/encoding/utf8</code></a> Returns page containing UTF-8 data.</li>
<li><a href="gzip" data-bare-link="true"><code>/gzip</code></a> Returns gzip-encoded data.</li>
<li><a href="deflate" data-bare-link="true"><code>/deflate</code></a> Returns deflate-encoded data.</li>
<li><del><a href="brotli" data-bare-link="true"><code>/brotli</code></a> Returns brotli-encoded data.</del> <i>Not implemented!</i></li>
<li><a href="status/418"><code>/status/:code</code></a> Returns given HTTP Status code.</li>
<li><a href="response-headers?Server=httpbin&amp;Content-Type=text%2Fplain%3B+charset%3DUTF-8"><code>/response-headers?key=val</code></a> Returns given response headers.</li>
<li><a href="redirect/6"><code>/redirect/:n</code></a> 302 Redirects <em>n</em> times.</li>
<li><a href="redirect-to?url=http%3A%2F%2Fexample.com%2F"><code>/redirect-to?url=foo</code></a> 302 Redirects to the <em>foo</em> URL.</li>
<li><a href="redirect-to?status_code=307&amp;url=http%3A%2F%2Fexample.com%2F"><code>/redirect-to?url=foo&status_code=307</code></a> 307 Redirects to the <em>foo</em> URL.</li>
<li><a href="relative-redirect/6"><code>/relative-redirect/:n</code></a> 302 Relative redirects <em>n</em> times.</li>
<li><a href="absolute-redirect/6"><code>/absolute-redirect/:n</code></a> 302 Absolute redirects <em>n</em> times.</li>
<li><a href="cookies" data-bare-link="true"><code>/cookies</code></a> Returns cookie data.</li>
<li><a href="cookies/set?k1=v1&amp;k2=v2"><code>/cookies/set?name=value</code></a> Sets one or more simple cookies.</li>
<li><a href="cookies/delete?k1=&amp;k2="><code>/cookies/delete?name</code></a> Deletes one or more simple cookies.</li>
<li><a href="basic-auth/user/passwd"><code>/basic-auth/:user/:passwd</code></a> Challenges HTTPBasic Auth.</li>
<li><a href="hidden-basic-auth/user/passwd"><code>/hidden-basic-auth/:user/:passwd</code></a> 404'd BasicAuth.</li>
<li><a href="digest-auth/auth/user/passwd/MD5"><code>/digest-auth/:qop/:user/:passwd/:algorithm</code></a> Challenges HTTP Digest Auth.</li>
<li><a href="digest-auth/auth/user/passwd/MD5"><code>/digest-auth/:qop/:user/:passwd</code></a> Challenges HTTP Digest Auth.</li>
<li><a href="stream/20"><code>/stream/:n</code></a> Streams <em>min(n, 100)</em> lines.</li>
<li><a href="delay/3"><code>/delay/:n</code></a> Delays responding for <em>min(n, 10)</em> seconds.</li>
<li><a href="drip?code=200&amp;numbytes=5&amp;duration=5"><code>/drip?numbytes=n&amp;duration=s&amp;delay=s&amp;code=code</code></a> Drips data over a duration after an optional initial delay, then (optionally) returns with the given status code.</li>
<li><a href="range/1024"><code>/range/1024?duration=s&amp;chunk_size=code</code></a> Streams <em>n</em> bytes, and allows specifying a <em>Range</em> header to select a subset of the data. Accepts a <em>chunk_size</em> and request <em>duration</em> parameter.</li>
<li><a href="html" data-bare-link="true"><code>/html</code></a> Renders an HTML Page.</li>
<li><a href="robots.txt" data-bare-link="true"><code>/robots.txt</code></a> Returns some robots.txt rules.</li>
<li><a href="deny" data-bare-link="true"><code>/deny</code></a> Denied by robots.txt file.</li>
<li><a href="cache" data-bare-link="true"><code>/cache</code></a> Returns 200 unless an If-Modified-Since or If-None-Match header is provided, when it returns a 304.</li>
<li><a href="etag/etag"><code>/etag/:etag</code></a> Assumes the resource has the given etag and responds to If-None-Match header with a 200 or 304 and If-Match with a 200 or 412 as appropriate.</li>
<li><a href="cache/60"><code>/cache/:n</code></a> Sets a Cache-Control header for <em>n</em> seconds.</li>
<li><a href="bytes/1024"><code>/bytes/:n</code></a> Generates <em>n</em> random bytes of binary data, accepts optional <em>seed</em> integer parameter.</li>
<li><a href="stream-bytes/1024"><code>/stream-bytes/:n</code></a> Streams <em>n</em> random bytes of binary data, accepts optional <em>seed</em> and <em>chunk_size</em> integer parameters.</li>
<li><a href="links/10"><code>/links/:n</code></a> Returns page containing <em>n</em> HTML links.</li>
<li><a href="image"><code>/image</code></a> Returns page containing an image based on sent Accept header.</li>
<li><a href="image/png"><code>/image/png</code></a> Returns a PNG image.</li>
<li><a href="image/jpeg"><code>/image/jpeg</code></a> Returns a JPEG image.</li>
<li><a href="image/webp"><code>/image/webp</code></a> Returns a WEBP image.</li>
<li><a href="image/svg"><code>/image/svg</code></a> Returns a SVG image.</li>
<li><a href="forms/post" data-bare-link="true"><code>/forms/post</code></a> HTML form that submits to <em>/post</em></li>
<li><a href="xml" data-bare-link="true"><code>/xml</code></a> Returns some XML</li>
</ul>

<h2 id="DESCRIPTION">DESCRIPTION</h2>

<p>Testing an HTTP Library can become difficult sometimes. <a href="http://requestb.in">RequestBin</a> is fantastic for testing POST requests, but doesn't let you control the response. This exists to cover all kinds of HTTP scenarios. Additional endpoints are being considered.</p>

<p>All endpoint responses are JSON-encoded.</p>

<h2 id="EXAMPLES">EXAMPLES</h2>

<h3 id="-curl-http-httpbin-org-ip">$ curl http://httpbin.org/ip</h3>

<pre><code>{"origin": "24.127.96.129"}
</code></pre>

<h3 id="-curl-http-httpbin-org-user-agent">$ curl http://httpbin.org/user-agent</h3>

<pre><code>{"user-agent": "curl/7.19.7 (universal-apple-darwin10.0) libcurl/7.19.7 OpenSSL/0.9.8l zlib/1.2.3"}
</code></pre>

<h3 id="-curl-http-httpbin-org-get">$ curl http://httpbin.org/get</h3>

<pre><code>{
   "args": {},
   "headers": {
      "Accept": "*/*",
      "Connection": "close",
      "Content-Length": "",
      "Content-Type": "",
      "Host": "httpbin.org",
      "User-Agent": "curl/7.19.7 (universal-apple-darwin10.0) libcurl/7.19.7 OpenSSL/0.9.8l zlib/1.2.3"
   },
   "origin": "24.127.96.129",
   "url": "http://httpbin.org/get"
}
</code></pre>

<h3 id="-curl-I-http-httpbin-org-status-418">$ curl -I http://httpbin.org/status/418</h3>

<pre><code>HTTP/1.1 418 I'M A TEAPOT
Server: nginx/0.7.67
Date: Mon, 13 Jun 2011 04:25:38 GMT
Connection: close
x-more-info: http://tools.ietf.org/html/rfc2324
Content-Length: 135
</code></pre>

<h3 id="-curl-https-httpbin-org-get-show_env-1">$ curl https://httpbin.org/get?show_env=1</h3>

<pre><code>{
  "headers": {
    "Content-Length": "",
    "Accept-Language": "en-US,en;q=0.8",
    "Accept-Encoding": "gzip,deflate,sdch",
    "X-Forwarded-Port": "443",
    "X-Forwarded-For": "109.60.101.240",
    "Host": "httpbin.org",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "User-Agent": "Mozilla/5.0 (X11; Linux i686) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.83 Safari/535.11",
    "X-Request-Start": "1350053933441",
    "Accept-Charset": "ISO-8859-1,utf-8;q=0.7,*;q=0.3",
    "Connection": "keep-alive",
    "X-Forwarded-Proto": "https",
    "Cookie": "_gauges_unique_day=1; _gauges_unique_month=1; _gauges_unique_year=1; _gauges_unique=1; _gauges_unique_hour=1",
    "Content-Type": ""
  },
  "args": {
    "show_env": "1"
  },
  "origin": "109.60.101.240",
  "url": "http://httpbin.org/get?show_env=1"
}
</code></pre>


<h2 id="AUTHOR">AUTHOR</h2>

<p>Ported to Go by <a href="https://github.com/mccutchen">Will McCutchen</a>.</p>
<p>From <a href="https://httpbin.org/">the original</a> <a href="http://kennethreitz.com/">Kenneth Reitz</a> project.</p>

<h2 id="SEE-ALSO">SEE ALSO</h2>

<p><a href="https://httpbin.org/">httpbin.org</a> &mdash; the original httpbin</p>

</div>

<a href="https://github.com/mccutchen/go-httpbin"><img style="position: absolute; top: 0; right: 0; border: 0;" src="https://camo.githubusercontent.com/38ef81f8aca64bb9a64448d0d70f1308ef5341ab/68747470733a2f2f73332e616d617a6f6e6177732e636f6d2f6769746875622f726962626f6e732f666f726b6d655f72696768745f6461726b626c75655f3132313632312e706e67" alt="Fork me on GitHub" data-canonical-src="https://s3.amazonaws.com/github/ribbons/forkme_right_darkblue_121621.png"></a>

</body>
</html>
`)
}

// IPHandler returns Origin IP.
func IPHandler(w http.ResponseWriter, r *http.Request) {
	h, _, _ := net.SplitHostPort(r.RemoteAddr)
	if err := writeJSON(w, ipResponse{h}); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json")) // TODO handle this error in writeJSON(w,v)
	}
}

// UserAgentHandler returns user agent.
func UserAgentHandler(w http.ResponseWriter, r *http.Request) {
	if err := writeJSON(w, userAgentResponse{r.UserAgent()}); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
	}
}

// HeadersHandler returns user agent.
func HeadersHandler(w http.ResponseWriter, r *http.Request) {
	if err := writeJSON(w, headersResponse{getHeaders(r)}); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
	}
}

// GetHandler returns user agent.
func GetHandler(w http.ResponseWriter, r *http.Request) {
	h, _, _ := net.SplitHostPort(r.RemoteAddr)

	v := getResponse{
		headersResponse: headersResponse{getHeaders(r)},
		ipResponse:      ipResponse{h},
		Args:            flattenValues(r.URL.Query()),
	}

	if err := writeJSON(w, v); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
	}
}

// PostHandler accept a post and echo its data back
func PostHandler(w http.ResponseWriter, r *http.Request) {
	h, _, _ := net.SplitHostPort(r.RemoteAddr)

	data, err := parseData(r)
	if err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to read body"))
		return
	}

	var jsonPayload interface{}
	if strings.Contains(r.Header.Get("Content-Type"), "json") {
		err := json.Unmarshal(data, &jsonPayload)
		if err != nil {
			writeErrorJSON(w, errors.Wrap(err, "failed to read body"))
			return
		}
	}

	v := postResponse{
		headersResponse: headersResponse{getHeaders(r)},
		ipResponse:      ipResponse{h},
		Args:            flattenValues(r.URL.Query()),
		Data:            string(data),
		JSON:            jsonPayload,
	}

	if err := writeJSON(w, v); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
	}
}

// RedirectHandler returns a 302 Found response if n=1 pointing
// to /get, otherwise to /redirect/(n-1)
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	n := mux.Vars(r)["n"]
	i, _ := strconv.Atoi(n) // shouldn't fail due to route pattern

	var loc string
	if i <= 1 {
		loc = "/get"
	} else {
		loc = fmt.Sprintf("/redirect/%d", i-1)
	}
	w.Header().Set("Location", loc)
	w.WriteHeader(http.StatusFound)
}

// AbsoluteRedirectHandler returns a 302 Found response if n=1 pointing
// to /host/get, otherwise to /host/absolute-redirect/(n-1)
func AbsoluteRedirectHandler(w http.ResponseWriter, r *http.Request) {
	n := mux.Vars(r)["n"]
	i, _ := strconv.Atoi(n) // shouldn't fail due to route pattern

	var loc string
	if i <= 1 {
		loc = "/get"
	} else {
		loc = fmt.Sprintf("/absolute-redirect/%d", i-1)
	}

	w.Header().Set("Location", "http://"+r.Host+loc)
	w.WriteHeader(http.StatusFound)
}

// RedirectToHandler returns a 302 Found response pointing to
// the url query parameter
func RedirectToHandler(w http.ResponseWriter, r *http.Request) {
	u := mux.Vars(r)["url"]
	w.Header().Set("Location", u)
	w.WriteHeader(http.StatusFound)
}

// StatusHandler returns a proper response for provided status code
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	code, _ := strconv.Atoi(mux.Vars(r)["code"])

	statusWritten := false
	switch code {
	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusUseProxy,
		http.StatusTemporaryRedirect:
		w.Header().Set("Location", "/redirect/1")
	case http.StatusUnauthorized: // 401
		w.Header().Set("WWW-Authenticate", `Basic realm="Fake Realm"`)
	case http.StatusPaymentRequired: // 402
		w.WriteHeader(code)
		statusWritten = true
		io.WriteString(w, "Fuck you, pay me!")
		w.Header().Set("x-more-info", "http://vimeo.com/22053820")
	case http.StatusNotAcceptable: // 406
		w.WriteHeader(code)
		statusWritten = true
		io.WriteString(w, `{"message": "Client did not request a supported media type.", "accept": ["image/webp", "image/svg+xml", "image/jpeg", "image/png", "image/*"]}`)
	case http.StatusTeapot:
		w.WriteHeader(code)
		statusWritten = true
		w.Header().Set("x-more-info", "http://tools.ietf.org/html/rfc2324")
		io.WriteString(w, `
    -=[ teapot ]=-

       _...._
     .'  _ _ '.
    | ."  ^  ". _,
    \_;'"---"'|//
      |       ;/
      \_     _/
        '"""'
`)
	}
	if !statusWritten {
		w.WriteHeader(code)
	}
}

// BytesHandler returns n random bytes of binary data and accepts an
// optional 'seed' integer query parameter.
func BytesHandler(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(mux.Vars(r)["n"]) // shouldn't fail due to route pattern

	seedStr := r.URL.Query().Get("seed")
	if seedStr == "" {
		seedStr = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	seed, _ := strconv.ParseInt(seedStr, 10, 64) // shouldn't fail due to route pattern
	rnd := rand.New(rand.NewSource(seed))
	buf := make([]byte, BinaryChunkSize)
	for n > 0 {
		rnd.Read(buf) // will never return err
		if n >= len(buf) {
			n -= len(buf)
			w.Write(buf)
		} else {
			// last chunk
			w.Write(buf[:n])
			break
		}
	}
}

// DelayHandler delays responding for min(n, 10) seconds and responds
// with /get endpoint
func DelayHandler(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.ParseFloat(mux.Vars(r)["n"], 64) // shouldn't fail due to route pattern

	// allow only millisecond precision
	duration := time.Millisecond * time.Duration(n*float64(time.Second/time.Millisecond))
	if duration > DelayMax {
		duration = DelayMax
	}
	time.Sleep(duration)
	GetHandler(w, r)
}

// StreamHandler writes a json object to a new line every second.
func StreamHandler(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(mux.Vars(r)["n"]) // shouldn't fail due to route pattern
	nl := []byte{'\n'}
	// allow only millisecond precision
	for i := 0; i < n; i++ {
		time.Sleep(StreamInterval)
		b, _ := json.Marshal(struct {
			N    int       `json:"n"`
			Time time.Time `json:"time"`
		}{i, time.Now().UTC()})
		w.Write(b)
		w.Write(nl)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

// CookiesHandler returns the cookies provided in the request.
func CookiesHandler(w http.ResponseWriter, r *http.Request) {
	if err := writeJSON(w, cookiesResponse{getCookies(r.Cookies())}); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
	}
}

// SetCookiesHandler sets the query key/value pairs as cookies
// in the response and returns a 302 redirect to /cookies.
func SetCookiesHandler(w http.ResponseWriter, r *http.Request) {
	for k := range r.URL.Query() {
		v := r.URL.Query().Get(k)
		http.SetCookie(w, &http.Cookie{
			Name:  k,
			Value: v,
			Path:  "/",
		})
	}
	w.Header().Set("Location", "/cookies")
	w.WriteHeader(http.StatusFound)
}

// DeleteCookiesHandler deletes cookies with provided query value keys
// in the response by settings a Unix epoch expiration date and returns
// a 302 redirect to /cookies.
func DeleteCookiesHandler(w http.ResponseWriter, r *http.Request) {
	for k := range r.URL.Query() {
		http.SetCookie(w, &http.Cookie{
			Name:    k,
			Value:   "",
			Path:    "/",
			Expires: time.Unix(0, 0),
			MaxAge:  0,
		})
	}
	w.Header().Set("Location", "/cookies")
	w.WriteHeader(http.StatusFound)
}

// DripHandler drips data over a duration after an optional initial delay,
// then optionally returns with the given status code.
func DripHandler(w http.ResponseWriter, r *http.Request) {
	var retCode int

	retCodeStr := r.URL.Query().Get("code")
	delayStr := r.URL.Query().Get("delay")
	durationSec, _ := strconv.ParseFloat(mux.Vars(r)["duration"], 32) // shouldn't fail due to route pattern
	numBytes, _ := strconv.Atoi(mux.Vars(r)["numbytes"])              // shouldn't fail due to route pattern

	if retCodeStr != "" { // optional: status code
		var err error
		retCode, err = strconv.Atoi(r.URL.Query().Get("code"))
		if err != nil {
			writeErrorJSON(w, errors.New("failed to parse 'code'"))
			return
		}
		w.WriteHeader(retCode)
	}

	if delayStr != "" { // optional: initial delay
		delaySec, err := strconv.ParseFloat(r.URL.Query().Get("delay"), 64)
		if err != nil {
			writeErrorJSON(w, errors.New("failed to parse 'delay'"))
			return
		}
		delayMs := (time.Second / time.Millisecond) * time.Duration(delaySec)
		time.Sleep(delayMs * time.Millisecond)
	}

	t := time.Second * time.Duration(durationSec) / time.Duration(numBytes)
	for i := 0; i < numBytes; i++ {
		w.Write([]byte{'*'})
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		time.Sleep(t)
	}
}

// CacheHandler returns 200 with the response of /get unless an If-Modified-Since
//or If-None-Match header is provided, when it returns a 304.
func CacheHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("If-Modified-Since") != "" || r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	GetHandler(w, r)
}

// SetCacheHandler sets a Cache-Control header for n seconds and returns with
// the /get response.
func SetCacheHandler(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(mux.Vars(r)["n"]) // shouldn't fail due to route pattern
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", n))
	GetHandler(w, r)
}

// GZIPHandler returns a GZIP-encoded response
func GZIPHandler(w http.ResponseWriter, r *http.Request) {
	h, _, _ := net.SplitHostPort(r.RemoteAddr)

	v := gzipResponse{
		headersResponse: headersResponse{getHeaders(r)},
		ipResponse:      ipResponse{h},
		Gzipped:         true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Content-Encoding", "gzip")
	ww := gzip.NewWriter(w)
	defer ww.Close() // flush
	if err := writeJSON(ww, v); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
	}
}

// DeflateHandler returns a DEFLATE-encoded response.
func DeflateHandler(w http.ResponseWriter, r *http.Request) {
	h, _, _ := net.SplitHostPort(r.RemoteAddr)

	v := deflateResponse{
		headersResponse: headersResponse{getHeaders(r)},
		ipResponse:      ipResponse{h},
		Deflated:        true,
	}

	w.Header().Set("Content-Encoding", "deflate")
	ww, _ := flate.NewWriter(w, flate.BestCompression)
	defer ww.Close() // flush
	if err := writeJSON(ww, v); err != nil {
		writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
	}
}

// RobotsTXTHandler returns a robots.txt response.
func RobotsTXTHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "User-agent: *\nDisallow: /deny\n")
}

// DenyHandler returns a plain-text response.
func DenyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, `
          .-''''''-.
        .' _      _ '.
       /   O      O   \
      :                :
      |                |
      :       __       :
       \  .-"'  '"-.  /
        '.          .'
          '-......-'
     YOU SHOULDN'T BE HERE
`)
}

// BasicAuthHandler challenges with given username and password.
func BasicAuthHandler(w http.ResponseWriter, r *http.Request) {
	basicAuthHandler(w, r, http.StatusUnauthorized)
}

// HiddenBasicAuthHandler challenges with given username and password and
// returns 404 if authentication fails.
func HiddenBasicAuthHandler(w http.ResponseWriter, r *http.Request) {
	basicAuthHandler(w, r, http.StatusNotFound)
}

func basicAuthHandler(w http.ResponseWriter, r *http.Request, status int) {
	user := mux.Vars(r)["u"]
	pass := mux.Vars(r)["p"]

	inUser, inPass, ok := r.BasicAuth()
	if !ok || inUser != user || inPass != pass {
		w.WriteHeader(status)
	} else {
		v := basicAuthResponse{
			Authenticated: true,
			User:          user,
		}
		if err := writeJSON(w, v); err != nil {
			writeErrorJSON(w, errors.Wrap(err, "failed to write json"))
		}
	}
}

// HTMLHandler returns some HTML response.
func HTMLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, htmlData)
}

// XMLHandler returns some XML response.
func XMLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, xmlData)
}

type circle struct {
	X, Y, R float64
}

func (c *circle) Brightness(x, y float64) uint8 {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		return 0
	}
	return 255
}

// GIFHandler returns an animated GIF image.
// Source: http://tech.nitoyon.com/en/blog/2016/01/07/go-animated-gif-gen/
func GIFHandler(rw http.ResponseWriter, r *http.Request) {
	var w, h int = 240, 240
	var hw, hh float64 = float64(w / 2), float64(h / 2)
	circles := []*circle{{}, {}, {}}

	var palette = []color.Color{
		color.RGBA{0x00, 0x00, 0x00, 0xff},
		color.RGBA{0x00, 0x00, 0xff, 0xff},
		color.RGBA{0x00, 0xff, 0x00, 0xff},
		color.RGBA{0x00, 0xff, 0xff, 0xff},
		color.RGBA{0xff, 0x00, 0x00, 0xff},
		color.RGBA{0xff, 0x00, 0xff, 0xff},
		color.RGBA{0xff, 0xff, 0x00, 0xff},
		color.RGBA{0xff, 0xff, 0xff, 0xff},
	}

	var images []*image.Paletted
	var delays []int
	steps := 20
	for step := 0; step < steps; step++ {
		img := image.NewPaletted(image.Rect(0, 0, w, h), palette)
		images = append(images, img)
		delays = append(delays, 0)

		θ := 2.0 * math.Pi / float64(steps) * float64(step)
		for i, circle := range circles {
			θ0 := 2 * math.Pi / 3 * float64(i)
			circle.X = hw - 40*math.Sin(θ0) - 20*math.Sin(θ0+θ)
			circle.Y = hh - 40*math.Cos(θ0) - 20*math.Cos(θ0+θ)
			circle.R = 50
		}

		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				img.Set(x, y, color.RGBA{
					circles[0].Brightness(float64(x), float64(y)),
					circles[1].Brightness(float64(x), float64(y)),
					circles[2].Brightness(float64(x), float64(y)),
					255,
				})
			}
		}
	}

	gif.EncodeAll(rw, &gif.GIF{
		Image: images,
		Delay: delays,
	})
}

// JPEGHandler returns a JPEG image.
func JPEGHandler(w http.ResponseWriter, r *http.Request) {
	jpeg.Encode(w, getImg(), nil)
}

// PNGHandler returns a PNG image.
func PNGHandler(w http.ResponseWriter, r *http.Request) {
	png.Encode(w, getImg())
}

func getImg() image.Image {
	const n = 512
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
	sq := func(i int) int { return i * i }

	for x := 0; x <= n; x++ {
		for y := 0; y <= n; y++ {
			if x == n/2 && y == n/2 {
				continue
			}
			d := math.Sqrt(float64(sq(abs(x-n/2)) + sq(abs(y-n/2))))
			if d > n/2 {
				continue
			}

			sin := float64(y-n/2) / d
			deg := math.Asin(sin)/math.Pi*359.0 + 180
			sec := int(deg) / 60

			var fix, mod *uint8
			var inc bool

			c := color.RGBA{0, 0, 0, 0xFF}
			switch sec {
			case 0:
				fix, mod = &c.R, &c.G
				inc = true
			case 1:
				fix, mod = &c.G, &c.R
				inc = false
			case 2:
				fix, mod = &c.G, &c.B
				inc = true
			case 3:
				fix, mod = &c.B, &c.G
				inc = false
			case 4:
				fix, mod = &c.B, &c.R
				inc = true
			case 5:
				fix, mod = &c.R, &c.B
				inc = false
			default:
				panic(fmt.Sprintf("deg=%f sec=%d", deg, sec))
			}

			v := uint8((int(deg) % 60) * 255.0 / 60.0)
			*fix = 255
			if inc {
				*mod = v
			} else {
				*mod = 255 - v
			}
			img.Set(x, y, c)

		}
	}
	return img
}

func parseData(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
