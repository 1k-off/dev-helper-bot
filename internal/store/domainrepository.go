package store

import (
	"database/sql"
	"github.com/souladm/dev-helper-bot/model"
	"time"
)

type DomainRepository struct {
	store *Store
}

func (r *DomainRepository) Create(d *model.Domain) error {
	_, err := r.store.db.Exec ("INSERT INTO domains(fqdn, ip, user_id, user_name, created_at, delete_at, basic_auth) VALUES (?, ?, ?, ?, ?, ?, ?)",
		d.FQDN, d.IP, d.UserId, d.UserName, d.CreatedAt, d.DeleteAt, d.BasicAuth)
	if err != nil {
		return err
	}
	return nil
}

func (r *DomainRepository) Get(userId string) (*model.Domain, error) {
	d := &model.Domain{}
	if err := r.store.db.QueryRow(
		"SELECT fqdn, ip, user_id, user_name, created_at, delete_at, basic_auth FROM domains WHERE user_id = ?", userId,
		).Scan(&d.FQDN, &d.IP, &d.UserId, &d.UserName, &d.CreatedAt, &d.DeleteAt, &d.BasicAuth); err == sql.ErrNoRows {
		return nil, errRecordNotFound
	}
	return d, nil
}

func (r *DomainRepository) Update (d *model.Domain) error {
	result, err := r.store.db.Exec(
		"UPDATE domains SET ip=?, delete_at=?, basic_auth=? WHERE user_id = ?", d.IP, d.DeleteAt, d.BasicAuth, d.UserId)
	if err != nil {
		return err
	}
	if ra, _ := result.RowsAffected(); ra == 0 {
		return errNotUpdated
	}
	return nil
}

func (r *DomainRepository) GetAllRecordsToDeleteInDays(days int) ([]model.Domain, error) {
	var domains []model.Domain

	planDeleteDate := time.Now().AddDate(0,0,days).Format("2006-01-02")
	rows, err := r.store.db.Query("SELECT fqdn, ip, user_id, user_name, created_at, delete_at, basic_auth FROM domains WHERE delete_at <= ?", planDeleteDate,)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		d := model.Domain{}
		if err := rows.Scan(&d.FQDN, &d.IP, &d.UserId, &d.UserName, &d.CreatedAt, &d.DeleteAt, &d.BasicAuth); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func (r *DomainRepository) DeleteByFqdn(fqdn string) error {
	res, err := r.store.db.Exec("DELETE from domains WHERE fqdn = ?", fqdn)
	if err != nil {
		return err
	}
	if ra, _ := res.RowsAffected(); ra < 1 {
		return errNoRowsAffected
	}
	return nil
}