package storage

import (
	"context"
	"fmt"

	"github.com/globusdigital/feature-toggles/toggle"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

const flagsCollection = "flags"

type Mongo struct {
	client *mongo.Client
	db     string
}

type flag struct {
	Name        string `bson:"name"`
	ServiceName string `bson:"serviceName"`

	RawValue string `bson:"rawValue"`
	Value    bool   `bson:"value"`
}

func NewMongo(ctx context.Context, url string) (*Mongo, error) {
	cs, err := connstring.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("parsing url: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, fmt.Errorf("connecting to mongo server: %v", err)
	}

	client.Database(cs.Database).Collection(flagsCollection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{"serviceName", 1}, {"name", 1}},
		Options: options.Index().SetUnique(true),
	})

	return &Mongo{client, cs.Database}, nil
}

func (s *Mongo) Get(ctx context.Context, serviceName string) ([]toggle.Flag, error) {
	coll := s.client.Database(s.db).Collection(flagsCollection)
	filter := bson.D{}
	if serviceName != "" {
		filter = bson.D{{"$or", bson.A{bson.D{{"serviceName", ""}}, bson.D{{"serviceName", serviceName}}}}}
	}
	c, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("getting flag data: %v", err)
	}

	var flags []flag
	if err := c.All(ctx, &flags); err != nil {
		return nil, fmt.Errorf("decoding flag data: %v", err)
	}

	ret := make([]toggle.Flag, len(flags))
	for i := range flags {
		ret[i] = toggle.Flag(flags[i])
	}

	return ret, nil
}

func (s *Mongo) Save(ctx context.Context, flags []toggle.Flag, initial bool) error {
	coll := s.client.Database(s.db).Collection(flagsCollection)

	models := make([]mongo.WriteModel, 0, len(flags))
	for _, f := range flags {
		if initial {
			res := coll.FindOne(ctx, bson.D{{"serviceName", f.ServiceName}, {"name", f.Name}}, options.FindOne().SetProjection(bson.D{{"_id", 1}}))
			if res.Err() == mongo.ErrNoDocuments {
				models = append(models, mongo.NewInsertOneModel().SetDocument(flag(f)))
			}
		} else {
			models = append(models, mongo.NewUpdateOneModel().
				SetUpsert(true).
				SetFilter(bson.D{{"serviceName", f.ServiceName}, {"name", f.Name}}).
				SetUpdate(bson.D{{"$set", flag(f)}}))
		}
	}

	if len(models) == 0 {
		return nil
	}

	if _, err := coll.BulkWrite(ctx, models); err != nil {
		return fmt.Errorf("writing flag data: %v", err)
	}

	return nil
}

func (s *Mongo) Delete(ctx context.Context, flags []toggle.Flag) error {
	coll := s.client.Database(s.db).Collection(flagsCollection)

	models := make([]mongo.WriteModel, 0, len(flags))

	for _, f := range flags {
		models = append(models, mongo.NewDeleteOneModel().
			SetFilter(bson.D{{"serviceName", f.ServiceName}, {"name", f.Name}}))
	}

	if len(models) == 0 {
		return nil
	}

	if _, err := coll.BulkWrite(ctx, models); err != nil {
		return fmt.Errorf("deleting flag data: %v", err)
	}

	return nil
}
