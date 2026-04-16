package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/imesart/apple-ads-cli/internal/api"
)

const (
	snapshotTestToken = "token"
	snapshotTestOrgID = "12345"
)

type snapshotRoundTripFunc func(*http.Request) (*http.Response, error)

func (f snapshotRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// HTTPSnapshot is a parsed normalized HTTP snapshot.
type HTTPSnapshot struct {
	Method  string
	URL     *url.URL
	Headers map[string]string
	Body    []byte
}

// NewSnapshotAPIClient returns an API client configured with stable test auth headers.
func NewSnapshotAPIClient() *api.Client {
	return api.NewClient(func(context.Context) (string, error) {
		return snapshotTestToken, nil
	}, snapshotTestOrgID, false)
}

// CaptureAPIRequestSnapshot executes req through the API client and returns its normalized snapshot.
func CaptureAPIRequestSnapshot(ctx context.Context, client *api.Client, req api.Request) (string, error) {
	var captured string
	client.SetHTTPClientForTesting(&http.Client{
		Transport: snapshotRoundTripFunc(func(r *http.Request) (*http.Response, error) {
			var (
				body []byte
				err  error
			)
			if r.Body != nil {
				body, err = io.ReadAll(r.Body)
				if err != nil {
					return nil, fmt.Errorf("reading request body: %w", err)
				}
			}
			captured, err = FormatHTTPRequestSnapshot(r.Method, r.URL, r.Header, body)
			if err != nil {
				return nil, err
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    r,
			}, nil
		}),
	})

	if err := client.Do(ctx, req, nil); err != nil {
		return "", err
	}
	return captured, nil
}

// FormatHTTPRequestSnapshot renders a normalized textual snapshot of an HTTP request.
func FormatHTTPRequestSnapshot(method string, u *url.URL, headers http.Header, body []byte) (string, error) {
	var buf strings.Builder
	buf.WriteString(method)
	buf.WriteString(" ")
	buf.WriteString(normalizeSnapshotURL(u).String())
	buf.WriteString("\n")

	normalizedHeaders := orderedSnapshotHeaders(u, headers)
	for _, line := range normalizedHeaders {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return buf.String(), nil
	}

	buf.WriteString("\n")
	formattedBody, err := formatSnapshotBody(body)
	if err != nil {
		return "", err
	}
	buf.Write(formattedBody)
	buf.WriteString("\n")
	return buf.String(), nil
}

// ParseHTTPSnapshot parses a normalized snapshot file.
func ParseHTTPSnapshot(data []byte) (*HTTPSnapshot, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return nil, fmt.Errorf("empty snapshot")
	}

	parts := strings.SplitN(text, "\n\n", 2)
	headerLines := strings.Split(parts[0], "\n")
	if len(headerLines) == 0 {
		return nil, fmt.Errorf("missing request line")
	}

	requestLine := strings.Fields(headerLines[0])
	if len(requestLine) != 2 {
		return nil, fmt.Errorf("invalid request line %q", headerLines[0])
	}
	parsedURL, err := url.Parse(requestLine[1])
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	headers := make(map[string]string)
	for _, line := range headerLines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header line %q", line)
		}
		headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	var body []byte
	if len(parts) == 2 {
		body = []byte(strings.TrimSpace(parts[1]))
	}

	return &HTTPSnapshot{
		Method:  requestLine[0],
		URL:     normalizeSnapshotURL(parsedURL),
		Headers: headers,
		Body:    body,
	}, nil
}

// ReadHTTPSnapshot reads and parses a snapshot file.
func ReadHTTPSnapshot(path string) (*HTTPSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseHTTPSnapshot(data)
}

// AssertGoldenSnapshot compares got with the golden file on disk.
func AssertGoldenSnapshot(t *testing.T, got, goldenPath string) {
	t.Helper()
	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden %s: %v", goldenPath, err)
	}
	want := strings.TrimRight(string(wantBytes), "\n")
	got = strings.TrimRight(got, "\n")
	if got != want {
		t.Fatalf("snapshot mismatch for %s\n\nwant:\n%s\n\ngot:\n%s", filepath.Base(goldenPath), want, got)
	}
}

func normalizeSnapshotURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	cloned := *u
	if cloned.RawQuery != "" {
		values, err := url.ParseQuery(cloned.RawQuery)
		if err == nil {
			cloned.RawQuery = values.Encode()
		}
	}
	return &cloned
}

func orderedSnapshotHeaders(u *url.URL, headers http.Header) []string {
	headerValue := func(key string) string {
		for existingKey, vals := range headers {
			if strings.EqualFold(existingKey, key) && len(vals) > 0 {
				return vals[0]
			}
		}
		return ""
	}

	values := map[string]string{}
	if got := headerValue("Accept"); got != "" {
		values["Accept"] = got
	}
	if got := headerValue("Authorization"); got != "" {
		values["Authorization"] = got
	}
	if got := headerValue("Content-Type"); got != "" {
		values["Content-Type"] = got
	}
	if u != nil && u.Host != "" {
		values["Host"] = u.Host
	}
	if got := headerValue("X-AP-Context"); got != "" {
		values["X-AP-Context"] = got
	}

	order := []string{"Accept", "Authorization", "Content-Type", "Host", "X-AP-Context"}
	lines := make([]string, 0, len(values))
	for _, key := range order {
		if value := values[key]; value != "" {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
		}
	}

	extras := make([]string, 0)
	for key, vals := range headers {
		skip := false
		for knownKey := range values {
			if strings.EqualFold(key, knownKey) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if len(vals) == 0 {
			continue
		}
		extras = append(extras, fmt.Sprintf("%s: %s", key, vals[0]))
	}
	sort.Strings(extras)
	lines = append(lines, extras...)
	return lines
}

func formatSnapshotBody(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return nil, nil
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err == nil {
		formatted, err := json.MarshalIndent(decoded, "", "  ")
		if err != nil {
			return nil, err
		}
		return formatted, nil
	}
	return body, nil
}
