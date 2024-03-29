package handlers

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/entities"
	"github.com/1k-off/dev-helper-bot/internal/webserver"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

const (
	timeStoreDomainWeek = 2
)

var (
	timeStoreDomain = timeStoreDomainWeek * time.Hour * 24 * 7
)

func (h *Handler) DomainCreate(userId, userName, ip string) (*entities.Domain, error) {
	if err := webserver.CheckIfIpAllowed(h.Webserver.AllowedSubnets, h.Webserver.DeniedIPs, ip); err != nil {
		return nil, err
	}

	fqdn := transformName(userName) + "." + h.Webserver.ParentDomain
	delDate := time.Now().Add(timeStoreDomain)
	deleteDate := time.Date(delDate.Year(), delDate.Month(), delDate.Day(), 9, 0, 0, delDate.Nanosecond(), delDate.Location())

	domain := &entities.Domain{
		FQDN:      fqdn,
		IP:        ip,
		UserId:    userId,
		UserName:  userName,
		CreatedAt: time.Now(),
		DeleteAt:  deleteDate,
		BasicAuth: true,
		FullSsl:   false,
		Port:      "80",
	}

	if err := h.Webserver.Service.Create(domain); err != nil {
		return nil, err
	}

	err := h.Store.DomainRepository().Create(domain)
	if err != nil {
		return nil, err
	}
	log.Info().Msg(fmt.Sprintf("[bot] created domain %s with IP %s. Scheduled delete date: %s.", domain.FQDN, domain.IP, domain.DeleteAt))
	return domain, nil
}

func (h *Handler) DomainUpdate(userId, param, value string) (string, error) {
	d, err := h.Store.DomainRepository().Get(userId)
	if err != nil {
		return "", err
	}

	updateExp := func() {
		delDate := time.Now().Add(timeStoreDomain)
		d.DeleteAt = time.Date(delDate.Year(), delDate.Month(), delDate.Day(), 9, 0, 0, delDate.Nanosecond(), delDate.Location())
	}

	switch param {
	case "":
		updateExp()
	case "expire":
		updateExp()
	case "ip":
		ip := value
		if err = webserver.CheckIfIpAllowed(h.Webserver.AllowedSubnets, h.Webserver.DeniedIPs, ip); err != nil {
			return "", err
		}
		d.IP = ip
		if err = h.updateNginxConf(d); err != nil {
			return "", err
		}
	case "basic-auth":
		ba, err := strconv.ParseBool(value)
		if err != nil {
			return "", err
		}
		d.BasicAuth = ba
		if err = h.updateNginxConf(d); err != nil {
			return "", err
		}
	case "full-ssl":
		fs, err := strconv.ParseBool(value)
		if err != nil {
			return "", err
		}
		d.FullSsl = fs
		if err = h.updateNginxConf(d); err != nil {
			return "", err
		}
	case "port":
		if value == "" {
			return "", fmt.Errorf("port can't be empty")
		}
		portInt, err := strconv.Atoi(value)
		if err != nil {
			return "", err
		}
		if portInt < 1 || portInt > 65535 {
			return "", fmt.Errorf("port must be in range 1-65535")
		}
		d.Port = value
		if err = h.updateNginxConf(d); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unknown parameter")
	}

	err = h.Store.DomainRepository().Update(d)
	if err != nil {
		return "", err
	}
	log.Info().Msg(fmt.Sprintf("[bot] updated domain %v", d))
	return "Updated", nil
}

func (h *Handler) updateNginxConf(d *entities.Domain) error {
	err := h.Webserver.Service.Delete(d.FQDN)
	if err != nil {
		return err
	}
	if err = h.Webserver.Service.Create(d); err != nil {
		return err
	}
	return nil
}

func (h *Handler) DomainDelete(userId string) (string, error) {
	d, err := h.Store.DomainRepository().Get(userId)
	if err != nil {
		return "", err
	}
	err = h.Store.DomainRepository().DeleteByFqdn(d.FQDN)
	if err != nil {
		return "", err
	}
	err = h.Webserver.Service.Delete(d.FQDN)
	if err != nil {
		return "", err
	}
	log.Info().Msg(fmt.Sprintf("[bot] deleted domain %v", d))
	return fmt.Sprintf("Deleted domain %s", d.FQDN), nil

}

// DomainGetExpired returns list of expired domains
func (h *Handler) DomainGetExpired() ([]*entities.Domain, error) {
	return h.Store.DomainRepository().GetAllRecordsToDeleteInDays(0)
}

// DomainGetExpirationSoon returns list of domains that will be deleted in 1 day
func (h *Handler) DomainGetExpirationSoon() ([]*entities.Domain, error) {
	return h.Store.DomainRepository().GetAllRecordsToDeleteInDays(1)
}

// DomainDeleteExpired deletes all expired domains
func (h *Handler) DomainDeleteExpired() error {
	var errors []error
	domains, err := h.DomainGetExpired()
	if err != nil {
		return err
	}
	for _, d := range domains {
		if err = h.Webserver.Service.Delete(d.FQDN); err != nil {
			log.Err(err).Msg(fmt.Sprintf("[bot] error deleting domain %v", d))
			errors = append(errors, err)
		}
		if err = h.Store.DomainRepository().DeleteByFqdn(d.FQDN); err != nil {
			log.Err(err).Msg(fmt.Sprintf("[bot] error deleting domain %v", d))
			errors = append(errors, err)
		}
		log.Info().Msg(fmt.Sprintf("[bot] deleted domain %v", d))
	}
	if len(errors) > 0 {
		return fmt.Errorf(fmt.Sprintf("one or more errors occured while deleting domains. %v", errors))
	}
	return nil
}
