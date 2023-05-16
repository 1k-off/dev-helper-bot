package pritunl

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

func (c *Client) CreateUserInDefaultOrg(login, email string) error {
	org, err := c.GetOrganization()
	if err != nil {
		return err
	}
	err = c.CreateUser(login, email, org.ID)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteUserFromDefaultOrg(email string) error {
	org, err := c.GetOrganization()
	if err != nil {
		return err
	}
	user, err := c.GetUserByEmail(email, org.ID)
	if err != nil {
		return err
	}
	err = c.DeleteUser(user.ID, org.ID)
	if err != nil {
		return err
	}
	return nil
}
