package store

import "github.com/1k-off/dev-helper-bot/model"

type DomainRepository interface {
	Create(d *model.Domain) error
	Get(userId string) (domain *model.Domain, err error)
	Update(domain *model.Domain) error
	GetAllRecordsToDeleteInDays(days int) (domains []*model.Domain, err error)
	DeleteByFqdn(fqdn string) error
}
