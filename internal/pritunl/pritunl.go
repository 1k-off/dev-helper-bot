package pritunl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Host         string
	Token        string
	Secret       string
	Organization string
	client       *http.Client
}

type Organization struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	AuthAPI    bool        `json:"auth_api"`
	AuthToken  interface{} `json:"auth_token"`
	AuthSecret interface{} `json:"auth_secret"`
	UserCount  int         `json:"user_count"`
}

type User struct {
	ID               string `json:"id"`
	Organization     string `json:"organization"`
	OrganizationName string `json:"organization_name"`
	Name             string `json:"name"`
	Email            string `json:"email"`
}

type Key struct {
	ID        string `json:"id"`
	KeyURL    string `json:"key_url"`
	KeyZipURL string `json:"key_zip_url"`
	KeyOncURL string `json:"key_onc_url"`
	ViewURL   string `json:"view_url"`
	URIURL    string `json:"uri_url"`
}

func New(host, token, secret, organization string) *Client {
	c := &http.Client{}
	return &Client{
		Host:         host,
		Token:        token,
		Secret:       secret,
		Organization: organization,
		client:       c,
	}
}

func (c *Client) GetOrganization() (org Organization, err error) {
	req, err := c.newRequest(http.MethodGet, endpointOrganization, nil)
	if err != nil {
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var organizations []Organization
	err = json.NewDecoder(resp.Body).Decode(&organizations)
	if err != nil {
		return
	}
	for _, organization := range organizations {
		if organization.Name == c.Organization {
			org = organization
		}
	}
	return org, nil
}

func (c *Client) GetUserByName(name string, orgId string) (user User, err error) {
	req, err := c.newRequest(http.MethodGet, endpointUser+"/"+orgId, nil)
	if err != nil {
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var users []User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return
	}
	for _, u := range users {
		if u.Name == name {
			user = u
		}
	}
	return user, nil
}
func (c *Client) GetUserByEmail(email string, orgId string) (user User, err error) {
	req, err := c.newRequest(http.MethodGet, endpointUser+"/"+orgId, nil)
	if err != nil {
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var users []User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return
	}
	for _, u := range users {
		if u.Email == email {
			user = u
		}
	}
	return user, nil
}

func (c *Client) GetUserKeys(userId, orgId string) (key Key, err error) {
	req, err := c.newRequest(http.MethodGet, endpointKey+"/"+orgId+"/"+userId, nil)
	if err != nil {
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&key)
	if err != nil {
		return
	}
	return key, nil
}

func (c *Client) GetUserKeyZipUrl(email string) (url string, err error) {
	org, err := c.GetOrganization()
	if err != nil {
		return
	}
	user, err := c.GetUserByEmail(email, org.ID)
	if err != nil {
		return
	}
	key, err := c.GetUserKeys(user.ID, org.ID)
	if err != nil {
		return
	}
	return c.Host + key.KeyZipURL, nil
}

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
