package store

import "github.com/1k-off/dev-helper-bot/internal/entities"

type DomainRepository interface {
	Create(d *entities.Domain) error
	Get(userId string) (domain *entities.Domain, err error)
	Update(domain *entities.Domain) error
	GetAllRecordsToDeleteInDays(days int) (domains []*entities.Domain, err error)
	DeleteByFqdn(fqdn string) error
}

type VPNEURepository interface {
	Create(vpnRecord *entities.VPNEU) error
	GetAllRecordsToDeactivateInMinutes(minutes int) (records []*entities.VPNEU, err error)
	SetInactive(record *entities.VPNEU) error
}
