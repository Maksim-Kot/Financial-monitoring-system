package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	appErrors "fms-project/internal/infrastructure/errors"

	"golang.org/x/time/rate"
)

type HttpClient struct {
	baseURL string
	hc      *http.Client
	headers http.Header

	rl      *rate.Limiter
	retries int
}

func (c *HttpClient) WithBaseURL(u string) *HttpClient {
	c.baseURL = strings.TrimRight(u, "/")
	return c
}

func (c *HttpClient) WithTimeout(d time.Duration) *HttpClient { c.hc.Timeout = d; return c }

func (c *HttpClient) WithTransport(tr http.RoundTripper) *HttpClient { c.hc.Transport = tr; return c }

func (c *HttpClient) WithHeader(k, v string) *HttpClient {
	clone := *c

	clone.headers = clone.headers.Clone()

	clone.headers.Add(k, v)

	return &clone
}

func (c *HttpClient) WithHeaders(h http.Header) *HttpClient {
	for k, vv := range h {
		for _, v := range vv {
			c.headers.Add(k, v)
		}
	}

	return c
}

func (c *HttpClient) WithRetries(n int) *HttpClient { c.retries = n; return c }

func (c *HttpClient) WithRateLimit(r rate.Limit, burst int) *HttpClient {
	c.rl = rate.NewLimiter(r, burst)
	return c
}

func NewHttpClient() *HttpClient {
	return &HttpClient{
		hc: &http.Client{
			Timeout:   15 * time.Second,
			Transport: http.DefaultTransport,
		},
		headers: make(http.Header),
		retries: 0,
	}
}

const (
	maxErrBody = 4 << 10 // 4 KiB
	maxJSON    = 1 << 20 // 1 MiB
)

func (c *HttpClient) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	for k, vv := range c.headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	tries := c.retries + 1
	var lastErr error

	for attempt := 0; attempt < tries; attempt++ {
		if c.rl != nil {
			if err := c.rl.Wait(ctx); err != nil {
				return nil, err
			}
		}

		attemptReq, err := prepareAttemptRequest(req, ctx)
		if err != nil {
			return nil, appErrors.Wrap(err, "http-adapter", fmt.Sprintf("prepare request for retry url:%v", req.URL))
		}

		resp, err := c.hc.Do(attemptReq)
		if err != nil {
			lastErr = err
			if !transientNetError(err) || attempt == tries-1 {
				break
			}
			if !sleepBackoff(ctx, attempt) {
				break
			}
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		b, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrBody))
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = appErrors.New("http-adapter", fmt.Sprintf("status %d: %s", resp.StatusCode, string(b)))
			if attempt < tries-1 && honorRetryAfter(ctx, resp.Header.Get("Retry-After")) {
				continue
			}
			if attempt < tries-1 {
				if !sleepBackoff(ctx, attempt) {
					break
				}
				continue
			}
			break
		}

		if resp.StatusCode >= 500 && resp.StatusCode != 501 && resp.StatusCode != 505 {
			lastErr = appErrors.New("http-adapter", fmt.Sprintf("status %d: %s", resp.StatusCode, string(b)))

			if attempt < tries-1 {
				if !sleepBackoff(ctx, attempt) {
					break
				}
				continue
			}
			break
		}

		lastErr = appErrors.New("http-adapter", fmt.Sprintf("status %d: %s", resp.StatusCode, string(b)))

		break
	}

	return nil, appErrors.Wrap(lastErr, "http-adapter", fmt.Sprintf("request failed url:%v", req.URL))
}

func prepareAttemptRequest(src *http.Request, ctx context.Context) (*http.Request, error) {
	clone := src.Clone(ctx)
	if src.Body != nil {
		if src.GetBody == nil {
			return nil, errors.New("request body cannot be replayed for retry")
		}
		body, err := src.GetBody()
		if err != nil {
			return nil, err
		}
		clone.Body = body
	}
	return clone, nil
}

func transientNetError(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) && (ne.Timeout()) {
		return true
	}
	// иные “connection reset by peer” и т.п.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "write: broken pipe") ||
		strings.Contains(msg, "temporary failure") ||
		strings.Contains(msg, "tls handshake timeout")
}

func honorRetryAfter(ctx context.Context, ra string) bool {
	ra = strings.TrimSpace(ra)
	if ra == "" {
		return false
	}
	sec, err := strconv.Atoi(ra)
	if err != nil || sec <= 0 {
		return false
	}
	t := time.NewTimer(time.Duration(sec) * time.Second)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func sleepBackoff(ctx context.Context, attempt int) bool {
	base := 100 * time.Millisecond
	backoff := time.Duration(1<<attempt) * base
	jitter := time.Duration(rand.Int63n(int64(50 * time.Millisecond)))
	t := time.NewTimer(backoff + jitter)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (c *HttpClient) abs(p string) string {
	if c.baseURL == "" {
		return p
	}
	return strings.TrimRight(c.baseURL, "/") + p
}

func (c *HttpClient) Get(ctx context.Context, p string) ([]byte, error) {
	url := c.abs(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, appErrors.Wrap(err, "http-adapter", fmt.Sprintf("build request for %v", url))
	}

	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, appErrors.Wrap(err, "http-adapter", fmt.Sprintf("read body for %v", url))
	}

	return body, nil
}

func (c *HttpClient) GetJSON(ctx context.Context, p string, out any) error {
	url := c.abs(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return appErrors.Wrap(err, "http-adapter", fmt.Sprintf("build request for %v", url))
	}

	resp, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	buf, _ := io.ReadAll(io.LimitReader(resp.Body, maxJSON))

	if err := json.Unmarshal(buf, out); err != nil {
		return appErrors.Wrap(err, "http-adapter",
			fmt.Sprintf("decode json for url: %v response: %s", url, truncateString(string(buf), 500)))
	}

	return nil
}

func (c *HttpClient) PostJSON(ctx context.Context, p string, in any, out any) error {
	url := c.abs(p)

	var bodyReader io.Reader = http.NoBody
	if in != nil {
		b, err := encodeJSONBody(in)
		if err != nil {
			return appErrors.Wrap(err, "http-adapter", fmt.Sprintf("encode json for %v", url))
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return appErrors.Wrap(err, "http-adapter", fmt.Sprintf("build request for %v", url))
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if out == nil {
		return nil
	}

	dec := json.NewDecoder(io.LimitReader(resp.Body, maxJSON))
	if err := dec.Decode(out); err != nil {
		return appErrors.Wrap(err, "http-adapter", fmt.Sprintf("decode json for %v", url))
	}

	return nil
}

func (c *HttpClient) PostJSONToStringResponse(ctx context.Context, p string, in any) (string, error) {
	url := c.abs(p)

	var bodyReader io.Reader = http.NoBody
	if in != nil {
		b, err := encodeJSONBody(in)
		if err != nil {
			return "", appErrors.Wrap(err, "http-adapter", fmt.Sprintf("encode json for %v", url))
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return "", appErrors.Wrap(err, "http-adapter", fmt.Sprintf("build request for %v", url))
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := c.do(ctx, req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", appErrors.Wrap(err, "http-adapter", fmt.Sprintf("read body for %v", url))
	}

	return string(b), nil
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func encodeJSONBody(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false) // ближе к JSON.stringify
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	b := bytes.TrimRight(buf.Bytes(), "\n") // убираем \n
	return b, nil
}
