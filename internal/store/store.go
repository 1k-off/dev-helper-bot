package store

import (
	"database/sql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type Store struct {
	db              *sql.DB
	DomainRepository *DomainRepository
}

func init() {
	if _, err := os.Stat("data/data.db"); os.IsNotExist(err) {
		os.Create("data/data.db")
	}
}

func New() (*Store, error) {
	db, err := sql.Open("sqlite3", "data/data.db")
	if err != nil {
		return nil, err
	}

	//driver, _ := sqlite3.WithInstance(db, &sqlite3.Config{})
	//m, err := migrate.NewWithDatabaseInstance(
	//	"file://migrations",
	//	"data", driver)
	//if err != nil {
	//	return nil, err
	//}
	//if err := m.Up(); err != nil {
	//	return nil, err
	//}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{
		db: db,
	}, nil
}

func (s *Store) Domain() *DomainRepository {
	if s.DomainRepository != nil {
		return s.DomainRepository
	}

	s.DomainRepository = &DomainRepository{
		store: s,
	}
	return s.DomainRepository
}