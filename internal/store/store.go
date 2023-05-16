package store

type Store interface {
	DomainRepository() DomainRepository
	Close() error
}
