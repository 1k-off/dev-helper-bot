package pritunl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (c *Client) newRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	timestamp := time.Now().Unix()
	nonce := strings.Replace(uuid.New().String(), "-", "", -1)
	authString := strings.Join([]string{c.Token, strconv.FormatInt(timestamp, 10), nonce, method, endpoint}, "&")

	h := hmac.New(sha256.New, []byte(c.Secret))
	h.Write([]byte(authString))
	signatureRaw := h.Sum(nil)
	signature := base64.StdEncoding.EncodeToString(signatureRaw)

	req, err := http.NewRequest(method, c.Host+endpoint, body)
	req.Header.Set("Auth-Token", c.Token)
	req.Header.Set("Auth-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("Auth-Nonce", nonce)
	req.Header.Set("Auth-Signature", signature)
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}
	return req, nil
}
