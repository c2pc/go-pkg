package service

import (
	"context"
	"errors"
	"strconv"

	"github.com/c2pc/go-pkg/v2/example/internal/i18n"
	"github.com/c2pc/go-pkg/v2/example/internal/model"
	"github.com/c2pc/go-pkg/v2/example/internal/repository"
	"github.com/c2pc/go-pkg/v2/task"
	model3 "github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"gorm.io/gorm"
)

var (
	ErrNewsNotFound    = apperr.New("news_not_found", apperr.WithTextTranslate(i18n.ErrNewsNotFound), apperr.WithCode(code.NotFound))
	ErrNewsListIsEmpty = apperr.New("news_list_is_empty", apperr.WithTextTranslate(i18n.ErrNewsListIsEmpty), apperr.WithCode(code.NotFound))
	ErrNewsExists      = apperr.New("news_exists_error", apperr.WithTextTranslate(i18n.ErrNewsExists), apperr.WithCode(code.AlreadyExists))
)

type INews interface {
	Trx(db *gorm.DB) INews
	List(ctx context.Context, m *model2.Meta[model.News]) error
	GetById(ctx context.Context, id int) (*model.News, error)
	Create(ctx context.Context, input NewsCreateInput) (*model.News, error)
	Update(ctx context.Context, id int, input NewsUpdateInput) error
	Delete(ctx context.Context, id int) error
}

type News struct {
	newsRepository repository.INewsRepository
}

func NewNews(
	newsRepository repository.INewsRepository,
) News {
	return News{
		newsRepository: newsRepository,
	}
}

func (s News) Trx(db *gorm.DB) INews {
	s.newsRepository = s.newsRepository.Trx(db)
	return s
}

func (s News) List(ctx context.Context, m *model2.Meta[model.News]) error {
	return s.newsRepository.Omit("content").Paginate(ctx, m, ``)
}

func (s News) GetById(ctx context.Context, id int) (*model.News, error) {
	news, err := s.newsRepository.Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrNewsNotFound
		}
		return nil, err
	}

	return news, nil
}

type NewsCreateInput struct {
	Title   string
	Content *string
}

func (s News) Create(ctx context.Context, input NewsCreateInput) (*model.News, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	news, err := s.newsRepository.Create(ctx, &model.News{
		Title:   input.Title,
		Content: input.Content,
		UserID:  userID,
	}, "id")
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrNewsExists
		}
		return nil, err
	}

	return news, nil
}

type NewsUpdateInput struct {
	Title   *string
	Content *string
}

func (s News) Update(ctx context.Context, id int, input NewsUpdateInput) error {
	news, err := s.newsRepository.Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrNewsNotFound
		}
		return err
	}

	var selects []interface{}
	if input.Title != nil && *input.Title != "" {
		news.Title = *input.Title
		selects = append(selects, "title")
	}
	if input.Content != nil {
		if *input.Content == "" {
			news.Content = nil
		} else {
			news.Content = input.Content
		}
		selects = append(selects, "content")
	}

	if len(selects) > 0 {
		if err = s.newsRepository.Update(ctx, news, selects, `id = ?`, news.ID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return ErrNewsExists
			}
			return err
		}
	}

	return nil
}

func (s News) Delete(ctx context.Context, id int) error {
	news, err := s.newsRepository.Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrNewsNotFound
		}
		return err
	}

	if err := s.newsRepository.Delete(ctx, `id = ?`, news.ID); err != nil {
		return err
	}

	return nil
}

type NewsExport struct {
	Title   string  `json:"title" csv:"title"`
	Content *string `json:"content" csv:"content"`
	UserID  int     `json:"user_id" csv:"user_id"`
}

type NewsExportInput struct {
	Filter model2.Filter `json:"filter"`
}

func (s News) Export(ctx context.Context, taskID int, data []byte, msgChan chan<- *model3.Message) (*model3.Message, error) {
	return task.Export[model.News, NewsExport, NewsExportInput](
		ctx,
		taskID,
		msgChan,
		data,
		ErrNewsListIsEmpty,
		func(nei NewsExportInput) error { return nil },
		func(ctx context.Context, input NewsExportInput) ([]model.News, error) {
			list, err := s.newsRepository.List(ctx, &input.Filter, "")
			if err != nil {
				return nil, err
			}
			return list, nil
		},
		func(item model.News) (NewsExport, error) {
			return NewsExport{
				Title:   item.Title,
				Content: item.Content,
				UserID:  item.UserID,
			}, nil
		},
	)
}

type NewsImportDataInput struct {
	Err     *string `json:"err,omitempty"`
	Title   string  `json:"title"`
	Content *string `json:"content"`
}

type NewsImportInput struct {
	UserID int                   `json:"user_id"`
	Data   []NewsImportDataInput `json:"data"`
}

func (s News) Import(ctx context.Context, taskID int, data []byte, msgChan chan<- *model3.Message) (*model3.Message, error) {
	return task.Import[NewsImportInput, NewsImportDataInput](
		ctx,
		taskID,
		msgChan,
		data,
		func(nii NewsImportInput) error { return nil },
		func(d NewsImportInput) []NewsImportDataInput {
			return d.Data
		},
		func(ctx context.Context, data NewsImportInput, input NewsImportDataInput) (key string, prevErr error, err error) {
			if input.Err != nil {
				prevErr = errors.New(*input.Err)
				return
			}

			ctx = mcontext.WithOpUserIDContext(ctx, data.UserID)

			err = s.newsRepository.DB().Transaction(func(tx *gorm.DB) error {
				resp, err := s.Trx(tx).Create(ctx, NewsCreateInput{
					Title:   input.Title,
					Content: input.Content,
				})
				if err != nil {
					return err
				}
				key = strconv.Itoa(resp.ID)
				return nil
			})

			return
		},
	)
}

type NewsMassUpdateInput struct {
	IDs     []int   `json:"ids"`
	Content *string `json:"content"`
}

func (s News) MassUpdate(ctx context.Context, taskID int, data []byte, msgChan chan<- *model3.Message) (*model3.Message, error) {
	return task.MassUpdate[NewsMassUpdateInput](
		ctx,
		taskID,
		msgChan,
		data,
		ErrNewsNotFound,
		func(nmui NewsMassUpdateInput) error { return nil },
		func(d NewsMassUpdateInput) []int {
			return d.IDs
		},
		func(ctx context.Context, ids []int) ([]int, error) {
			pluckedIDs, err := s.newsRepository.PluckIDs(ctx, `id IN (?)`, ids)
			if err != nil {
				return nil, err
			}
			return pluckedIDs, nil
		},
		func(ctx context.Context, id int, input NewsMassUpdateInput) error {
			return s.newsRepository.DB().Transaction(func(tx *gorm.DB) error {
				err := s.Trx(tx).Update(ctx, id, NewsUpdateInput{Content: input.Content})
				if err != nil {
					return err
				}
				return nil
			})
		},
	)
}

type NewsMassDeleteInput struct {
	IDs []int `json:"ids"`
}

func (s News) MassDelete(ctx context.Context, taskID int, data []byte, msgChan chan<- *model3.Message) (*model3.Message, error) {
	return task.MassDelete[NewsMassDeleteInput](
		ctx,
		taskID,
		msgChan,
		data,
		ErrNewsNotFound,
		func(nmdi NewsMassDeleteInput) error { return nil },
		func(d NewsMassDeleteInput) []int {
			return d.IDs
		},
		func(ctx context.Context, ids []int) ([]int, error) {
			pluckedIDs, err := s.newsRepository.PluckIDs(ctx, `id IN (?)`, ids)
			if err != nil {
				return nil, err
			}
			return pluckedIDs, nil
		},
		func(ctx context.Context, input NewsMassDeleteInput, id int) error {
			return s.newsRepository.DB().Transaction(func(tx *gorm.DB) error {
				err := s.Trx(tx).Delete(ctx, id)
				if err != nil {
					return err
				}
				return nil
			})
		},
	)
}
