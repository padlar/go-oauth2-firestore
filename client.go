package fstore

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"google.golang.org/api/iterator"
	"gopkg.in/oauth2.v3/models"
	"time"
)

type store struct {
	c *firestore.Client
	n string // Top-level collection name.
	t time.Duration
}

func (s *store) Put(token *models.Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.t)
	defer cancel()
	_, _, err := s.c.Collection(s.n).Add(ctx, token)
	return err
}

func (s *store) Get(key string, val interface{}) (*models.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.t)
	defer cancel()
	iter := s.c.Collection(s.n).Where(key, "==", val).Limit(1).Documents(ctx)
	doc, err := first(iter)
	if err != nil {
		return nil, err
	}
	info := &models.Token{}
	err = doc.DataTo(info)
	return info, err
}

func (s *store) Del(key string, val interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.t)
	defer cancel()
	return s.c.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		query := s.c.Collection(s.n).Where(key, "==", val).Limit(1)
		iter := tx.Documents(query)
		doc, err := first(iter)
		if err != nil {
			return err
		}
		return tx.Delete(doc.Ref)
	})
}

// ErrDocumentDoesNotExist is returned whenever a Firestore document does not exist.
var ErrDocumentDoesNotExist = errors.New("document does not exist")

func first(iter *firestore.DocumentIterator) (*firestore.DocumentSnapshot, error) {
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, ErrDocumentDoesNotExist
	}
	if err != nil {
		return nil, err
	}
	return doc, nil
}
