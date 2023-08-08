package handlers

func (h *Handler) VpnGetConfigUrl(email string) (string, error) {
	url, err := h.PritunlClient.GetUserKeyZipUrl(email)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (h *Handler) VpnCreateUser(login, email string) error {
	err := h.PritunlClient.CreateUserInDefaultOrg(login, email)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) VpnDeleteUser(email string) error {
	err := h.PritunlClient.DeleteUserFromDefaultOrg(email)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) GetVpnWelcomeMessage() string {
	return h.MessageTemplates["vpnWelcomeMessage"]
}
