package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionName = "pessoas"
	dbName         = "rinha"
)

type MongoDB struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func MustNewMongoDB() *MongoDB {
	// Conectando com o DB.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://root:rootpassword@host.docker.internal:27017"))
	if err != nil {
		panic(fmt.Errorf("error connecting to db: %w", err))
	}

	// Testando conexão.
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(fmt.Errorf("error pinging db: %w", err))
	}

	collection := client.Database(dbName).Collection(collectionName)
	collection.Drop(context.TODO()) // removing all previously stored documents.
	defLanguage := "portuguese"
	defUnique := true
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "id", Value: 1}},
			Options: &options.IndexOptions{
				Unique: &defUnique,
			},
		}, {
			Keys: bson.D{{Key: "nome", Value: "text"}, {Key: "apelido", Value: "text"}, {Key: "stack", Value: "text"}},
			Options: &options.IndexOptions{
				DefaultLanguage: &defLanguage,
			},
		}}
	if _, err := collection.Indexes().CreateMany(context.TODO(), indexes); err != nil {
		panic(fmt.Errorf("error creating db index: %w", err))
	}
	return &MongoDB{
		client:     client,
		collection: collection,
	}
}

func (db *MongoDB) Create(p *Pessoa) error {
	_, err := db.collection.InsertOne(context.Background(), p)
	if err != nil {
		return fmt.Errorf("error inserting pessoa: %w", err)
	}
	return err
}

func (db *MongoDB) Get(id string) (*Pessoa, error) {
	p := new(Pessoa)
	err := db.collection.FindOne(context.Background(), bson.D{{Key: "id", Value: id}}).Decode(p)
	switch err {
	case nil:
		return p, nil
	case mongo.ErrNoDocuments:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("error decoding get result for term (%s): %w", id, err)
	}
}

func (db *MongoDB) Search(term string) ([]*Pessoa, error) {
	filter := bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: term}}}}
	opts := options.Find().SetLimit(50)
	cursor, err := db.collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error searching for term (%s): %w", term, err)
	}
	defer cursor.Close(context.TODO())

	var results []*Pessoa
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, fmt.Errorf("error decoding desarch results for term (%s): %w", term, err)
	}
	return results, nil
}

func (db *MongoDB) Count() (int, error) {
	count, err := db.collection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, fmt.Errorf("error counting pessoas: %w", err)
	}
	return int(count), nil
}
