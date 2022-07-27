package internal

import (
	"fmt"
	"strconv"
	"time"

	"github.com/couchbase/gocb/v2"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

type Store struct {
	internalCollection  *gocb.BinaryCollection
	stretchesCollection *gocb.Collection
}

func NewStore() (*Store, error) {
	bucket, err := connectToCluster()
	if err != nil {
		return nil, fmt.Errorf("could not connect to cluster: %w", err)
	}

	settingUp := false
	if err = setupScopesAndCollections(bucket, &settingUp); err != nil {
		return nil, fmt.Errorf("could not setup scopes and collections: %w", err)
	}

	scope := bucket.Scope(viper.GetString("scope"))
	internalCollection := scope.Collection("internal").Binary()
	stretchesCollection := scope.Collection("stretches")

	_, err = internalCollection.Increment("next-id", &gocb.IncrementOptions{Initial: 1})
	if err != nil {
		return nil, fmt.Errorf("could not increment next-id: %w", err)
	}

	return &Store{internalCollection: internalCollection, stretchesCollection: stretchesCollection}, nil
}

func (s *Store) GetID() (string, error) {
	id, err := s.internalCollection.Increment("next-id", &gocb.IncrementOptions{Delta: 1})
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(id.Content() - 1, 10), nil
}

func (s *Store) Insert(id string, stretch *Stretch) error {
	_, err := s.stretchesCollection.Insert(id, stretch, nil)
	if err != nil {
		return fmt.Errorf("could not insert stretch: %w", err)
	}

	return nil
}

func (s *Store) Upsert(id string, stretch *Stretch) error {
	_, err := s.stretchesCollection.Upsert(id, stretch, nil)
	if err != nil {
		return fmt.Errorf("could not upsert stretch: %w", err)
	}

	return nil
}

func connectToCluster() (*gocb.Bucket, error) {
	cluster, err := gocb.Connect(viper.GetString("connection"), gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{
			Username: viper.GetString("username"),
			Password: viper.GetString("password"),
		},
		SecurityConfig: gocb.SecurityConfig{TLSSkipVerify: true},
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to cluster: %w", err)
	}

	bucket := cluster.Bucket(viper.GetString("bucket"))
	err = bucket.WaitUntilReady(time.Second, nil)
	if err != nil {
		return nil, fmt.Errorf("could not get bucket: %w", err)
	}

	return bucket, nil
}

func setupScopesAndCollections(bucket *gocb.Bucket, settingUp *bool) error {
	cm := bucket.Collections()

	scopes, err := cm.GetAllScopes(nil)
	if err != nil {
		return fmt.Errorf("could not get scopes: %w", err)
	}

	var collections []gocb.CollectionSpec

	scopeName := viper.GetString("scope")

	i := slices.IndexFunc(scopes, func(s gocb.ScopeSpec) bool { return s.Name == scopeName })
	if i == -1 {
		*settingUp = true
		fmt.Println("Setting up database")

		if err = cm.CreateScope(scopeName, nil); err != nil {
			return fmt.Errorf("could not create '%s' scope: %w", scopeName, err)
		}
	} else {
		collections = scopes[i].Collections
	}

	for _, collection := range []string{"internal", "stretches"} {
		if i := slices.IndexFunc(collections, func(c gocb.CollectionSpec) bool { return c.Name == collection }); i == -1 {
			if !*settingUp {
				*settingUp = true
				fmt.Println("Setting up database")
			}

			err = cm.CreateCollection(gocb.CollectionSpec{ScopeName: scopeName, Name: collection}, nil)
			if err != nil {
				return fmt.Errorf("could not create collection: %w", err)
			}
		}
	}

	return nil
}
