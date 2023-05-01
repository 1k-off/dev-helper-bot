package mongostore

import (
	"context"
	"devhelper/internal/store"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type DataStore struct {
	client           *mongo.Client
	db               *mongo.Database
	ctx              context.Context
	domainRepository *domainRepository
}

func New(uri string) *DataStore {
	cs, err := connstring.ParseAndValidate(uri)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	return &DataStore{
		client: client,
		db:     client.Database(cs.Database),
		ctx:    ctx,
	}
}
func (s *DataStore) DomainRepository() store.DomainRepository {
	if s.domainRepository != nil {
		return s.domainRepository
	}

	c := s.db.Collection(store.DomainCollection)
	_, err := c.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: store.DomainUserIdKey, Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("")
	}
	s.domainRepository = &domainRepository{
		store:      s,
		collection: c,
	}
	return s.domainRepository
}

func (s *DataStore) Close() error {
	return s.client.Disconnect(s.ctx)
}
