package storage

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/globusdigital/feature-toggles/toggle"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongo_Get(t *testing.T) {
	type args struct {
		ctx         context.Context
		serviceName string
	}
	tests := []struct {
		name    string
		args    args
		initial []toggle.Flag
		want    []toggle.Flag
	}{
		{name: "no data", args: args{ctx: context.Background()}},
		{name: "data", args: args{ctx: context.Background()}, initial: initialData, want: initialData},
		{name: "data - svc1", args: args{ctx: context.Background(), serviceName: "svc1"}, initial: initialData, want: []toggle.Flag{
			initialData[0], initialData[1], initialData[3], initialData[4],
		}},
		{name: "data - svc2", args: args{ctx: context.Background(), serviceName: "svc2"}, initial: initialData, want: initialData[2:]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, cleanup := getTempDB(t)
			defer cleanup()

			a := assert.New(t)
			s, err := NewMongo(tt.args.ctx, url)
			a.NoError(err)

			if tt.initial != nil {
				err := s.Save(context.Background(), tt.initial, false)
				a.NoError(err)
			}

			got, err := s.Get(tt.args.ctx, tt.args.serviceName)
			sort.Slice(got, func(i, j int) bool {
				return got[i].Name < got[j].Name
			})

			a.NoError(err)
			a.Equal(tt.want, got)
		})
	}
}

func TestMongo_Save(t *testing.T) {
	type args struct {
		ctx     context.Context
		flags   []toggle.Flag
		initial bool
	}
	tests := []struct {
		name     string
		args     args
		initial  []toggle.Flag
		expected []toggle.Flag
	}{
		{name: "save data", args: args{ctx: context.Background(), flags: initialData}, expected: initialData},
		{name: "save data after initial", args: args{ctx: context.Background(), flags: []toggle.Flag{
			{Name: "n2", ServiceName: "svc1", RawValue: "1", Value: true},
			{Name: "n3", ServiceName: "svc1", RawValue: "0"},
		}, initial: true}, initial: initialData, expected: []toggle.Flag{
			initialData[0], initialData[1], initialData[2], initialData[3], initialData[4],
			{Name: "n3", ServiceName: "svc1", RawValue: "0"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, cleanup := getTempDB(t)
			defer cleanup()

			a := assert.New(t)
			s, err := NewMongo(tt.args.ctx, url)
			a.NoError(err)

			if tt.initial != nil {
				err := s.Save(context.Background(), tt.initial, false)
				a.NoError(err)
			}

			err = s.Save(tt.args.ctx, tt.args.flags, tt.args.initial)
			a.NoError(err)
			c, err := s.client.Database(s.db).Collection(flagsCollection).Find(context.Background(), bson.D{})
			a.NoError(err)

			var data []flag
			err = c.All(context.Background(), &data)
			a.NoError(err)

			a.Len(data, len(tt.expected))

			for _, v := range data {
				var found bool
				for _, e := range tt.expected {
					if e == toggle.Flag(v) {
						found = true
					}
				}
				a.True(found, "expected %#v", v)
			}
		})
	}
}

func TestMongo_Delete(t *testing.T) {
	type args struct {
		ctx   context.Context
		flags []toggle.Flag
	}
	tests := []struct {
		name     string
		args     args
		initial  []toggle.Flag
		expected []toggle.Flag
	}{
		{name: "no data", args: args{ctx: context.Background(), flags: initialData[3:]}, expected: nil},
		{name: "save data", args: args{ctx: context.Background(), flags: initialData[3:]}, initial: initialData, expected: initialData[:3]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, cleanup := getTempDB(t)
			defer cleanup()

			a := assert.New(t)
			s, err := NewMongo(tt.args.ctx, url)
			a.NoError(err)

			if tt.initial != nil {
				err := s.Save(context.Background(), tt.initial, false)
				a.NoError(err)
			}

			err = s.Delete(tt.args.ctx, tt.args.flags)
			a.NoError(err)
			c, err := s.client.Database(s.db).Collection(flagsCollection).Find(context.Background(), bson.D{})
			a.NoError(err)

			var data []flag
			err = c.All(context.Background(), &data)
			a.NoError(err)

			a.Len(data, len(tt.expected))

			for _, v := range data {
				var found bool
				for _, e := range tt.expected {
					if e == toggle.Flag(v) {
						found = true
					}
				}
				a.True(found, "expected %#v", v)
			}
		})
	}
}

var mongoURL = "mongodb://localhost:27017/"

func getTempDB(t *testing.T) (string, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	db := fmt.Sprintf("test_%d", time.Now().Unix())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL+db))
	if err != nil {
		t.Skipf("Mongodb connection error: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Mongodb ping error: %v", err)
	}

	client.Disconnect(ctx)

	return mongoURL + db, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL+db))
		if err != nil {
			t.Fatalf("Error connecting to Mongo: %v", err)
		}

		if err := client.Database(db).Drop(ctx); err != nil {
			t.Fatalf("Error dropping database %s: %v", db, err)
		}
	}
}
