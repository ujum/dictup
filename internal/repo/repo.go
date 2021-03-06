package repo

import (
	"context"
	"github.com/ujum/dictap/internal/client"
	"github.com/ujum/dictap/internal/config"
	"github.com/ujum/dictap/internal/domain"
	"github.com/ujum/dictap/internal/repo/mongo"
	"github.com/ujum/dictap/pkg/logger"
)

const (
	usersCollection      = "users"
	wordsCollection      = "words"
	wordgroupsCollection = "wordgroups"
)

type Repositories struct {
	UserRepo      UserRepo
	WordRepo      WordRepo
	WordGroupRepo WordGroupRepo
}

type UserRepo interface {
	FindByUID(ctx context.Context, uid string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (string, error)
	FindAll(ctx context.Context) ([]*domain.User, error)
	DeleteByUID(ctx context.Context, uid string) error
	Update(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type WordRepo interface {
	Create(ctx context.Context, word *domain.Word) (string, error)
	FindByGroup(ctx context.Context, groupID string) ([]*domain.Word, error)
	FindByName(ctx context.Context, name string) (*domain.Word, error)
	AddToGroup(ctx context.Context, wordID string, groupID string) error
	FindByNameAndGroup(ctx context.Context, wordName string, groupID string) (*domain.Word, error)
	RemoveFromGroup(ctx context.Context, name string, groupID string) error
}

type WordGroupRepo interface {
	Create(ctx context.Context, word *domain.WordGroup) (string, error)
	FindByIDAndUser(ctx context.Context, groupID string, userUID string) (*domain.WordGroup, error)
	FindByLangAndUser(ctx context.Context, langBinding *domain.LangBinding, userUID string, def bool) (*domain.WordGroup, error)
	FindAllByLangAndUser(ctx context.Context, langBinding *domain.LangBinding, userUID string) ([]*domain.WordGroup, error)
}

func New(cfg *config.Config, log logger.Logger, clients *client.Clients) *Repositories {
	mongoDatabase := clients.Mongo.Client.Database(cfg.Datasource.Mongo.Database)
	return &Repositories{
		UserRepo:      mongo.NewUserRepoMongo(log, mongoDatabase.Collection(usersCollection)),
		WordRepo:      mongo.NewWordRepoMongo(log, mongoDatabase.Collection(wordsCollection)),
		WordGroupRepo: mongo.NewWordGroupRepoMongo(log, mongoDatabase.Collection(wordgroupsCollection)),
	}
}
