package createuserservice

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/lexizz/cumloys/internal/pkg/logger"
	"github.com/lexizz/cumloys/internal/pkg/utils"
	"github.com/lexizz/cumloys/internal/repository"
	"github.com/lexizz/cumloys/internal/service"
)

var _ service.CreateUserServiceInterface = &createUserService{}

var (
	ErrUserExists     = errors.New("user has already exists")
	ErrUserCreation   = errors.New("failed creation user, this user wasn't created")
	ErrGenerationHash = errors.New("field generation password hash")
)

type createUserService struct {
	userRepository repository.UserRepositoryInterface
	logger         logger.Logger
}

func New(userRepository repository.UserRepositoryInterface, logger logger.Logger) *createUserService {
	return &createUserService{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (service *createUserService) Handle(ctx context.Context, newLogin string, newPwd string) (*uuid.UUID, error) {
	if len(newLogin) < 1 || len(newPwd) < 1 {
		return nil, errors.New("field login or password are empty")
	}

	isUserExists, err := service.userRepository.IsExists(ctx, newLogin)
	if err != nil {
		return nil, err
	}

	if isUserExists {
		return nil, ErrUserExists
	}

	passwordHash, err := utils.GeneratePasswordHash(newPwd)
	if err != nil {
		return nil, ErrGenerationHash
	}

	lastInsertID, errInsert := service.userRepository.Insert(ctx, newLogin, passwordHash)
	if errInsert != nil {
		return nil, ErrUserCreation
	}

	return lastInsertID, nil
}
