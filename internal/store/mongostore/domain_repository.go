package mongostore

import (
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/store"
	"github.com/1k-off/dev-helper-bot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type domainRepository struct {
	store      *DataStore
	collection *mongo.Collection
}

func (r *domainRepository) Create(domain *model.Domain) error {
	// TODO validation
	res, err := r.collection.InsertOne(r.store.ctx, domain)
	if err != nil {
		r.store.log.Debug().Msg(fmt.Sprintf("[database] tried to create record: %v", domain))
		r.store.log.Error().Err(err).Msg("")
		return err
	}
	domain.Id = res.InsertedID.(primitive.ObjectID).Hex()
	r.store.log.Info().Msg(fmt.Sprintf("[database] created record: %s", domain.FQDN))
	r.store.log.Debug().Msg(fmt.Sprintf("[database] created record: %v", domain))
	return nil
}
func (r *domainRepository) Get(userId string) (domain *model.Domain, err error) {
	filter := bson.M{store.DomainUserIdKey: userId}
	result := r.collection.FindOne(r.store.ctx, filter)
	err = result.Decode(&domain)
	if err != nil {
		return nil, err
	}
	return domain, nil
}
func (r *domainRepository) Update(domain *model.Domain) error {
	// TODO validation
	filter := bson.D{{store.DomainUserIdKey, domain.UserId}}
	update := bson.D{{"$set",
		bson.D{
			{store.DomainIpKey, domain.IP},
			{store.DomainBasicAuthKey, domain.BasicAuth},
			{store.DomainFullSslKey, domain.FullSsl},
			{store.DomainDeleteAtKey, domain.DeleteAt},
		},
	}}

	result, err := r.collection.UpdateOne(r.store.ctx, filter, update)
	if err != nil {
		r.store.log.Error().Err(err).Msg("")
		return err
	}
	if result.MatchedCount == 0 {
		r.store.log.Debug().Msg(fmt.Sprintf("[database] tried to update record: %v", domain))
		r.store.log.Error().Msg("[database] no records found")
		return store.ErrRecordNotFound
	}
	if result.ModifiedCount == 0 {
		r.store.log.Debug().Msg(fmt.Sprintf("[database] tried to update record: %v", domain))
		r.store.log.Error().Msg("[database] no records modified")
		return store.ErrNoRowsUpdated
	}
	r.store.log.Info().Msg(fmt.Sprintf("[database] updated record: %s", domain.FQDN))
	r.store.log.Debug().Msg(fmt.Sprintf("[database] updated record: %v", domain))
	return nil
}
func (r *domainRepository) GetAllRecordsToDeleteInDays(days int) (domains []*model.Domain, err error) {
	result, err := r.collection.Find(r.store.ctx, bson.M{store.DomainDeleteAtKey: bson.M{
		"$lte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, days)),
	}})
	defer result.Close(r.store.ctx)
	if err != nil {
		r.store.log.Error().Err(err)
		r.store.log.Debug().Msg("[database] error when trying to find records to delete by time expiration")
		return nil, err
	}
	for result.Next(r.store.ctx) {
		var d *model.Domain
		_ = result.Decode(&d)
		domains = append(domains, d)
	}
	return domains, nil
}

func (r *domainRepository) DeleteByFqdn(fqdn string) error {
	filter := bson.D{{store.DomainFqdnKey, fqdn}}
	opts := options.Delete().SetCollation(&options.Collation{})
	result, err := r.collection.DeleteOne(r.store.ctx, filter, opts)
	if err != nil {
		r.store.log.Error().Err(err).Msg("")
		return err
	}
	if result.DeletedCount == 0 {
		r.store.log.Debug().Msg(fmt.Sprintf("[database] tried to delete record with fqdn: %s", fqdn))
		r.store.log.Error().Msg("[database] no records deleted")
		return store.ErrNoRowsDeleted
	}
	r.store.log.Info().Msg(fmt.Sprintf("[database] deleted record with fqdn: %s", fqdn))
	return nil
}
