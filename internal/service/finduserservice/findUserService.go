package finduserservice

import (
	"context"
	"errors"

	"github.com/lexizz/cumloys/internal/models"
	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/repository"
)

type findUserService struct {
	userRepository repository.UserRepositoryInterface
	logger         logger.Logger
}

var (
	ErrInternal     = errors.New("internal error")
	ErrUserNotFound = errors.New("user not found")
)

func New(userRepository repository.UserRepositoryInterface, logger logger.Logger) *findUserService {
	return &findUserService{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (service *findUserService) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	user, errQuery := service.userRepository.GetUserByLogin(ctx, login)
	if errQuery != nil {
		return nil, ErrInternal
	}

	if user != nil {
		return user, nil
	}

	return nil, ErrUserNotFound
}
