package store

type Store interface {
	DomainRepository() DomainRepository
	VPNEURepository() VPNEURepository
	Close() error
}
