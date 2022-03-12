package store

type Data interface {
	DomainRepository() DomainRepository
	Close() error
}
