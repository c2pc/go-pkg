package service

import (
	"github.com/c2pc/go-pkg/v2/example/internal/repository"
)

type Services struct {
	NewsService INewsService
}

type Deps struct {
	Repositories repository.Repositories
}

func NewServices(deps Deps) Services {
	return Services{
		NewsService: NewNewsService(deps.Repositories.NewsRepository),
	}
}
