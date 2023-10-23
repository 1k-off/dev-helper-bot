package pritunl

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
	Disabled         bool   `json:"disabled"`
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
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
	if user.ID == "" {
		return user, errors.New("user not found")
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
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
	if user.ID == "" {
		return user, errors.New("user not found")
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	err = json.NewDecoder(resp.Body).Decode(&key)
	if err != nil {
		return
	}
	return key, nil
}

func (c *Client) CreateUser(login, email, orgId string) error {
	u, _ := c.GetUserByEmail(email, orgId)
	if u.ID != "" {
		return errors.New("user already exists")
	}
	u, _ = c.GetUserByName(login, orgId)
	if u.ID != "" {
		return errors.New("user already exists")
	}
	user := &User{
		Name:         login,
		Email:        email,
		Organization: orgId,
	}
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}
	req, err := c.newRequest(http.MethodPost, endpointUser+"/"+orgId, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	var createdUsers []User
	err = json.NewDecoder(resp.Body).Decode(&createdUsers)
	if err != nil {
		return err
	}
	if createdUsers[0].ID == "" {
		return errors.New("user id is empty")
	}
	return nil
}

func (c *Client) DeleteUser(id, orgId string) error {
	req, err := c.newRequest(http.MethodDelete, endpointUser+"/"+orgId+"/"+id, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	return nil
}

func (c *Client) UpdateUser(id, orgId string, userInfo *User) error {
	payload, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}
	req, err := c.newRequest(http.MethodPut, endpointUser+"/"+orgId+"/"+id, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	return nil
}

func (c *Client) ActivateUser(email, orgId string) error {
	user, err := c.GetUserByEmail(email, orgId)
	if err != nil {
		return err
	}
	user.Disabled = false
	err = c.UpdateUser(user.ID, orgId, &user)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeactivateUser(email, orgId string) error {
	user, err := c.GetUserByEmail(email, orgId)
	if err != nil {
		return err
	}
	user.Disabled = true
	err = c.UpdateUser(user.ID, orgId, &user)
	if err != nil {
		return err
	}
	return nil
}
