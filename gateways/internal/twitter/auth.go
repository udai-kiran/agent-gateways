package twitter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// CRCToken computes the CRC response token for Twitter webhook verification.
func CRCToken(crcToken, consumerSecret string) string {
	mac := hmac.New(sha256.New, []byte(consumerSecret))
	mac.Write([]byte(crcToken))
	return "sha256=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// SignRequest adds an OAuth 1.0a Authorization header to the request.
func SignRequest(req *http.Request, cfg Config, body string) {
	params := map[string]string{
		"oauth_consumer_key":     cfg.APIKey,
		"oauth_nonce":            generateNonce(),
		"oauth_signature_method": "HMAC-SHA256",
		"oauth_timestamp":        fmt.Sprintf("%d", time.Now().Unix()),
		"oauth_token":            cfg.AccessToken,
		"oauth_version":          "1.0",
	}

	// Collect all parameters for signature base string.
	sigParams := make(map[string]string)
	for k, v := range params {
		sigParams[k] = v
	}

	// Build sorted parameter string.
	var keys []string
	for k := range sigParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", percentEncode(k), percentEncode(sigParams[k])))
	}
	paramString := strings.Join(paramPairs, "&")

	// Build signature base string.
	baseURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	baseString := fmt.Sprintf("%s&%s&%s",
		strings.ToUpper(req.Method),
		percentEncode(baseURL),
		percentEncode(paramString),
	)

	// Sign.
	signingKey := percentEncode(cfg.APISecret) + "&" + percentEncode(cfg.AccessSecret)
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	params["oauth_signature"] = signature

	// Build Authorization header.
	var headerPairs []string
	for k, v := range params {
		headerPairs = append(headerPairs, fmt.Sprintf(`%s="%s"`, k, percentEncode(v)))
	}
	req.Header.Set("Authorization", "OAuth "+strings.Join(headerPairs, ", "))
}

func percentEncode(s string) string {
	return url.QueryEscape(s)
}

func generateNonce() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
