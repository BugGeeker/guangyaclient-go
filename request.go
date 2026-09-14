package guangyaclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) accountHeaders() http.Header {
	h := http.Header{
		"Accept": {"*/*"}, "Content-Type": {"application/json"},
		"Origin": {"https://www.guangyapan.com"}, "Referer": {"https://www.guangyapan.com/"},
		"User-Agent": {userAgent}, "X-Client-Id": {clientID}, "X-Client-Version": {"0.0.1"},
		"X-Device-Id": {c.DeviceID}, "X-Device-Model": {"chrome%2F147.0.0.0"}, "X-Device-Name": {"PC-Chrome"},
		"X-Device-Sign": {"wdi10." + c.DeviceID + randomHex(16)}, "X-Net-Work-Type": {"NONE"},
		"X-Os-Version": {"MacIntel"}, "X-Platform-Version": {"1"}, "X-Protocol-Version": {"301"},
		"X-Provider-Name": {"NONE"}, "X-Sdk-Version": {"9.0.2"},
	}
	return h
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := io.ReadFull(strings.NewReader(strings.Repeat("0", n)), b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", time.Now().UnixNano())[:min(n*2, 16)]
}

func (c *Client) do(method, rawURL string, body []byte, headers http.Header, query url.Values) (*http.Response, error) {
	c.mu.Lock()
	expired := c.RefreshTokenValue != "" && !c.TokenExpiresAt.IsZero() && !time.Now().Before(c.TokenExpiresAt)
	c.mu.Unlock()
	if expired {
		if _, err := c.RefreshToken(""); err != nil {
			return nil, err
		}
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, values := range headers {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}
	if req.Header.Get("Traceparent") == "" {
		req.Header.Set("Traceparent", GenerateTraceparent())
	}
	if !strings.HasPrefix(rawURL, accountBase) {
		req.Header.Set("Accept", "application/json, text/plain, */*")
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("DID", c.DeviceID)
		req.Header.Set("DT", "4")
		req.Header.Set("Origin", "https://www.guangyapan.com")
		req.Header.Set("Referer", "https://www.guangyapan.com/")
		req.Header.Set("User-Agent", userAgent)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized && c.RefreshTokenValue != "" {
		resp.Body.Close()
		if _, err := c.RefreshToken(""); err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Traceparent", GenerateTraceparent())
		resp, err = c.HTTP.Do(req)
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("guangya: HTTP %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return resp, nil
}

func request[T any](c *Client, method, endpoint string, payload any, headers http.Header) (T, error) {
	var result T
	body, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}
	resp, err := c.do(method, endpoint, body, headers, nil)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func publicPost[T any](c *Client, path string, payload any) (T, error) {
	headers := http.Header{
		"Accept": {"application/json, text/plain, */*"}, "Content-Type": {"application/json"},
		"DID": {GenerateDID()}, "DT": {"4"}, "Origin": {"https://www.guangyapan.com"},
		"Referer": {"https://www.guangyapan.com/"}, "User-Agent": {userAgent},
	}
	return request[T](c, "POST", apiBase+path, payload, headers)
}
