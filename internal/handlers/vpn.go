package handlers

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/entities"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

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

func (h *Handler) VpnGetEUConfigUrl(login, email, userId string, durationHours int) (string, error) {
	deactivateAtOffset := time.Hour
	debug := os.Getenv("OOOPS_DEBUG")
	if debug == "true" {
		deactivateAtOffset = time.Minute
	}
	var vpnEuRecord *entities.VPNEU
	if durationHours != 1 && durationHours != 2 && durationHours != 4 {
		return "", fmt.Errorf("duration must be 1, 2 or 4 hours")
	}
	org, err := h.PritunlEUClient.GetOrganization()
	if err != nil {
		return "", err
	}
	user, err := h.PritunlEUClient.GetUserByEmail(email, org.ID)
	if err != nil {
		if err.Error() == "user not found" {
			err = h.PritunlEUClient.CreateUser(login, email, org.ID)
			if err != nil {
				return "", err
			}
			user, err = h.PritunlEUClient.GetUserByEmail(email, org.ID)
			vpnEuRecord = &entities.VPNEU{
				UserName:     login,
				UserEmail:    email,
				UserId:       userId,
				CreatedAt:    time.Now(),
				DeactivateAt: time.Now().Add(deactivateAtOffset * time.Duration(durationHours)),
				Active:       true,
			}
			err = h.Store.VPNEURepository().Create(vpnEuRecord)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	if user.Disabled {
		err = h.PritunlEUClient.ActivateUser(email, org.ID)
		if err != nil {
			return "", err
		}
		vpnEuRecord = &entities.VPNEU{
			UserName:     login,
			UserEmail:    email,
			UserId:       userId,
			CreatedAt:    time.Now(),
			DeactivateAt: time.Now().Add(deactivateAtOffset * time.Duration(durationHours)),
			Active:       true,
		}
		err = h.Store.VPNEURepository().Create(vpnEuRecord)
		if err != nil {
			return "", err
		}
	}
	url, err := h.PritunlEUClient.GetUserKeyZipUrl(email)
	if err != nil {
		return "", err
	}
	hours := "hours"
	if durationHours == 1 {
		hours = "hour"
	}
	result := fmt.Sprintf("Here is your VPN config for %v %s:\n%v\nYou may use your existing profile if you already have it.", durationHours, hours, url)
	return result, nil
}

// VPNEUGetExpired returns list of expired VPN accounts on EU server
func (h *Handler) VPNEUGetExpired() ([]*entities.VPNEU, error) {
	return h.Store.VPNEURepository().GetAllRecordsToDeactivateInMinutes(0)
}

// VPNEUGetExpirationSoon returns list of VPN accounts on EU server that will be deleted soon
func (h *Handler) VPNEUGetExpirationSoon() ([]*entities.VPNEU, error) {
	return h.Store.VPNEURepository().GetAllRecordsToDeactivateInMinutes(10)
}

// VPNEUDeactivateExpired deactivates all expired vpn accounts on the EU server
func (h *Handler) VPNEUDeactivateExpired() error {
	var errors []error
	org, err := h.PritunlEUClient.GetOrganization()
	if err != nil {
		return err
	}
	accounts, err := h.VPNEUGetExpired()
	if err != nil {
		return err
	}
	for _, a := range accounts {
		if err = h.PritunlEUClient.DeactivateUser(a.UserEmail, org.ID); err != nil {
			log.Err(err).Msg(fmt.Sprintf("[bot] error deactivating user %v", a))
			errors = append(errors, err)
		}
		if err = h.Store.VPNEURepository().SetInactive(a); err != nil {
			log.Err(err).Msg(fmt.Sprintf("[bot] error deactivating user %v", a))
			errors = append(errors, err)
		}
		log.Info().Msg(fmt.Sprintf("[bot] deactivated user %v", a))
	}
	if len(errors) > 0 {
		return fmt.Errorf(fmt.Sprintf("one or more errors occured while deactivating users: %v", errors))
	}
	return nil
}
