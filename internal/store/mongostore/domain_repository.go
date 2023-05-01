package mongostore

import (
	"context"
	"fmt"
	"github.com/1k-off/dev-helper-bot/internal/entities"
	"github.com/1k-off/dev-helper-bot/internal/store"
	"github.com/rs/zerolog/log"
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

func (r *domainRepository) Create(domain *entities.Domain) error {
	// TODO validation
	res, err := r.collection.InsertOne(r.store.ctx, domain)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("[database] tried to create record: %v", domain))
		log.Error().Err(err).Msg("")
		return err
	}
	domain.Id = res.InsertedID.(primitive.ObjectID).Hex()
	log.Info().Msg(fmt.Sprintf("[database] created record: %s", domain.FQDN))
	log.Debug().Msg(fmt.Sprintf("[database] created record: %v", domain))
	return nil
}
func (r *domainRepository) Get(userId string) (domain *entities.Domain, err error) {
	filter := bson.M{store.DomainUserIdKey: userId}
	result := r.collection.FindOne(r.store.ctx, filter)
	err = result.Decode(&domain)
	if err != nil {
		return nil, err
	}
	return domain, nil
}
func (r *domainRepository) Update(domain *entities.Domain) error {
	// TODO validation
	filter := bson.D{{store.DomainUserIdKey, domain.UserId}}
	update := bson.D{{"$set",
		bson.D{
			{store.DomainIpKey, domain.IP},
			{store.DomainBasicAuthKey, domain.BasicAuth},
			{store.DomainFullSslKey, domain.FullSsl},
			{store.DomainDeleteAtKey, domain.DeleteAt},
			{store.DomainPortKey, domain.Port},
		},
	}}

	result, err := r.collection.UpdateOne(r.store.ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	if result.MatchedCount == 0 {
		log.Debug().Msg(fmt.Sprintf("[database] tried to update record: %v", domain))
		log.Error().Msg("[database] no records found")
		return store.ErrRecordNotFound
	}
	if result.ModifiedCount == 0 {
		log.Debug().Msg(fmt.Sprintf("[database] tried to update record: %v", domain))
		log.Error().Msg("[database] no records modified")
		return store.ErrNoRowsUpdated
	}
	log.Info().Msg(fmt.Sprintf("[database] updated record: %s", domain.FQDN))
	log.Debug().Msg(fmt.Sprintf("[database] updated record: %v", domain))
	return nil
}
func (r *domainRepository) GetAllRecordsToDeleteInDays(days int) (domains []*entities.Domain, err error) {
	result, err := r.collection.Find(r.store.ctx, bson.M{store.DomainDeleteAtKey: bson.M{
		"$lte": primitive.NewDateTimeFromTime(time.Now().AddDate(0, 0, days)),
	}})
	defer func(result *mongo.Cursor, ctx context.Context) {
		err := result.Close(ctx)
		if err != nil {
			log.Error().Err(err)
			log.Debug().Msg("[database] error when trying to close cursor")
		}
	}(result, r.store.ctx)
	if err != nil {
		log.Error().Err(err)
		log.Debug().Msg("[database] error when trying to find records to delete by time expiration")
		return nil, err
	}
	for result.Next(r.store.ctx) {
		var d *entities.Domain
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
		log.Error().Err(err).Msg("")
		return err
	}
	if result.DeletedCount == 0 {
		log.Debug().Msg(fmt.Sprintf("[database] tried to delete record with fqdn: %s", fqdn))
		log.Error().Msg("[database] no records deleted")
		return store.ErrNoRowsDeleted
	}
	log.Info().Msg(fmt.Sprintf("[database] deleted record with fqdn: %s", fqdn))
	return nil
}
