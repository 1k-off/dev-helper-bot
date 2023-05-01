package handlers

import (
	"devhelper/internal/entities"
	"devhelper/internal/nginx"
	"fmt"
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

func (h *Handler) DomainCreate(userId, userName, ip string) (string, error) {
	if err := nginx.CheckIfIpAllowed(h.NginxConfig.AllowedSubnets, h.NginxConfig.DeniedIPs, ip); err != nil {
		return "", err
	}

	fqdn := transformName(userName) + "." + h.NginxConfig.ParentDomain
	deleteDate := time.Now().Add(timeStoreDomain)
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

	if err := nginx.Create(domain); err != nil {
		return "", err
	}

	err := h.Store.DomainRepository().Create(domain)
	if err != nil {
		return "", err
	}
	log.Info().Msg(fmt.Sprintf("[bot] created domain %s with IP %s. Scheduled delete date: %s.", domain.FQDN, domain.IP, domain.DeleteAt))
	return fmt.Sprintf("Created domain %s with IP %s. Scheduled delete date: %s.", domain.FQDN, domain.IP, domain.DeleteAt.Format("2006-01-02 15:04:05")), nil
}

func (h *Handler) DomainUpdate(userId, param, value string) (string, error) {
	d, err := h.Store.DomainRepository().Get(userId)
	if err != nil {
		return "", err
	}

	updateExp := func() {
		d.DeleteAt = time.Now().Add(timeStoreDomain)
	}

	switch param {
	case "":
		updateExp()
	case "expire":
		updateExp()
	case "ip":
		ip := value
		if err = nginx.CheckIfIpAllowed(h.NginxConfig.AllowedSubnets, h.NginxConfig.DeniedIPs, ip); err != nil {
			return "", err
		}
		d.IP = ip
		if err = updateNginxConf(d); err != nil {
			return "", err
		}
	case "basic-auth":
		ba, err := strconv.ParseBool(value)
		if err != nil {
			return "", err
		}
		d.BasicAuth = ba
		if err = updateNginxConf(d); err != nil {
			return "", err
		}
	case "full-ssl":
		fs, err := strconv.ParseBool(value)
		if err != nil {
			return "", err
		}
		d.FullSsl = fs
		if err = updateNginxConf(d); err != nil {
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
		if err = updateNginxConf(d); err != nil {
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

func updateNginxConf(d *entities.Domain) error {
	err := nginx.Delete(d.FQDN)
	if err != nil {
		return err
	}
	if err = nginx.Create(d); err != nil {
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
	err = nginx.Delete(d.FQDN)
	if err != nil {
		return "", err
	}
	log.Info().Msg(fmt.Sprintf("[bot] deleted domain %v", d))
	return fmt.Sprintf("Deleted domain %s", d.FQDN), nil

}
