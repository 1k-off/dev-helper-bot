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
	"time"
)

type vpnEuRepository struct {
	store      *DataStore
	collection *mongo.Collection
}

func (r *vpnEuRepository) Create(vpnRecord *entities.VPNEU) error {
	// TODO validation
	res, err := r.collection.InsertOne(r.store.ctx, vpnRecord)
	if err != nil {
		log.Debug().Msg(fmt.Sprintf("[database] tried to create record: %v", vpnRecord))
		log.Error().Err(err).Msg("")
		return err
	}
	vpnRecord.Id = res.InsertedID.(primitive.ObjectID).Hex()
	log.Info().Msg(fmt.Sprintf("[database] created record: %s", vpnRecord.UserEmail))
	log.Debug().Msg(fmt.Sprintf("[database] created record: %v", vpnRecord))
	return nil
}

func (r *vpnEuRepository) GetAllRecordsToDeactivateInMinutes(minutes int) (records []*entities.VPNEU, err error) {
	result, err := r.collection.Find(r.store.ctx, bson.M{
		"$and": []bson.M{
			{
				store.VpnEUDeactivateAt: bson.M{
					"$lte": primitive.NewDateTimeFromTime(time.Now().Add(time.Minute * time.Duration(minutes))),
				},
			},
			{
				store.VpnEUActive: true,
			},
		},
	})
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
		var r *entities.VPNEU
		_ = result.Decode(&r)
		records = append(records, r)
	}
	return records, nil
}

func (r *vpnEuRepository) SetInactive(record *entities.VPNEU) error {
	filter := bson.D{{store.VpnEuUserEmail, record.UserEmail}, {store.VpnEUActive, true}}
	update := bson.D{{"$set",
		bson.D{
			{store.VpnEUActive, false},
		},
	}}

	result, err := r.collection.UpdateOne(r.store.ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	log.Debug().Msg(fmt.Sprintf("[database] updated record: %v", result))
	return nil
}

func (r *vpnEuRepository) GetActiveByUserEmail(userEmail string) (bool, error) {
	var record *entities.VPNEU
	err := r.collection.FindOne(r.store.ctx, bson.M{store.VpnEuUserEmail: userEmail, store.VpnEUActive: true}).Decode(&record)
	if err != nil {
		return false, err
	}
	return true, nil
}
