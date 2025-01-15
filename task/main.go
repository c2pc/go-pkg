package task

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/c2pc/go-pkg/v2/sse"
	model3 "github.com/c2pc/go-pkg/v2/task/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/c2pc/go-pkg/v2/task/internal/logger"
	"github.com/c2pc/go-pkg/v2/task/internal/model"
	"github.com/c2pc/go-pkg/v2/task/internal/repository"
	"github.com/c2pc/go-pkg/v2/task/internal/runner"
	"github.com/c2pc/go-pkg/v2/task/internal/service"
	"github.com/c2pc/go-pkg/v2/task/internal/transport/api/handler"
	"github.com/c2pc/go-pkg/v2/task/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

type Consumers map[string]Consumer

type Consumer interface {
	service.Consumer
}

type Queue interface {
	Run(data runner.Data)
	Stop(id int)
	TaskResults() chan runner.TaskResult
}

type Tasker interface {
	InitHandler(secured *gin.RouterGroup, unsecured *gin.RouterGroup)
	ExportHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc
	ImportHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc
	MassUpdateHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc
	MassDeleteHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc
}

type Task struct {
	debug       string
	db          *gorm.DB
	handler     handler.IHandler
	runner      Queue
	taskService service.ITaskService
}

type Config struct {
	Debug       string
	DB          *gorm.DB
	Transaction mw.ITransaction
	Services    Consumers
	TokenString string
	SseSvc      sse.SSE
}

func NewTask(ctx context.Context, cfg Config) (Tasker, error) {
	queue := runner.NewRunner(ctx, cfg.Debug)

	repositories := repository.NewRepositories(cfg.DB)

	consumers := make(service.Consumers, len(cfg.Services))
	for name, consumer := range cfg.Services {
		consumers[name] = service.Consumer(consumer)
	}

	if cfg.TokenString == "" {
		return nil, apperr.New("tokenString is required")
	}

	taskService := service.NewTaskService(repositories.TaskRepository, consumers, queue, cfg.TokenString, cfg.SseSvc)

	handlers := handler.NewHandlers(taskService, cfg.Transaction)

	exporter := &Task{
		handler:     handlers,
		runner:      queue,
		taskService: taskService,
		db:          cfg.DB,
		debug:       cfg.Debug,
	}

	go exporter.listen(ctx)

	if err := exporter.reset(ctx); err != nil {
		return nil, err
	}

	return exporter, nil
}

func (e *Task) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-e.runner.TaskResults():
			ctx2 := context.WithValue(ctx, constant.OperationID, fmt.Sprintf("runner-task-%d", result.ID))

			var input service.TaskUpdateInput
			var status string
			if result.Status != nil {
				status = e.getStatus(*result.Status)
				input.Status = &status
			}

			msg := model3.NewMessage()
			var appError apperr.Error
			if result.Error != nil {
				if !errors.As(result.Error, &appError) {
					appError = apperr.ErrInternal.WithError(result.Error)
				}
				msg.SetError(apperr.Translate(appError, translator.EN.String()))
				input.Output = msg
			} else {
				input.Output = result.Message
			}

			err := e.taskService.Update(ctx2, result.ID, input)

			if level.Is(e.debug, level.TEST) {
				logger.LogInfo(ctx2, "TYPE - %s | ID - %d | STATUS - %s | NAME - %s | CLIENT_ID - %d | TASK_ERROR - %v | UPDATE_ERROR - %v",
					result.Type, result.ID, status, result.Name, result.ClientID, appError, err)
			}
		}
	}
}

func (e *Task) reset(ctx context.Context) error {
	list, err := e.taskService.GetFull(ctx, nil, model.StatusRunning)
	if err != nil {
		return err
	}

	if len(list) > 0 {
		var ids []int
		for _, t := range list {
			ids = append(ids, t.ID)
		}

		err = e.taskService.UpdateStatus(ctx, model.StatusPending, ids...)
		if err != nil {
			return err
		}
	}

	err = e.taskService.RunTasks(ctx, []string{model.StatusPending})
	if err != nil {
		return err
	}

	return nil
}

func (e *Task) InitHandler(secured *gin.RouterGroup, unsecured *gin.RouterGroup) {
	e.handler.Init(secured, unsecured)
}

func (e *Task) NewTask(c *gin.Context, tp string, name string, data []byte) (*model.Task, error) {
	var task *model.Task
	err := e.db.Transaction(func(tx *gorm.DB) error {
		var err error
		task, err = e.taskService.Trx(tx).Create(c.Request.Context(), service.TaskCreateInput{
			Name: name,
			Type: tp,
			Data: data,
		})
		if err != nil {
			tx.Rollback()
			return err
		}

		err = e.taskService.Trx(tx).RunTasks(c.Request.Context(), []string{}, task.ID)
		if err != nil {
			tx.Rollback()
			return err
		}

		return err
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (e *Task) getStatus(status string) string {
	switch status {
	case runner.StatusRunning:
		return model.StatusRunning
	case runner.StatusStopped:
		return model.StatusStopped
	case runner.StatusCompleted:
		return model.StatusSuccess
	case runner.StatusFailed:
		return model.StatusFailed
	}
	return model.StatusPending
}

func (e *Task) ExportHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		cred, err := bind(c)
		if err != nil {
			response.Response(c, err)
			return
		}

		task, err := e.NewTask(c, model3.Export, name, cred)
		if err != nil {
			response.Response(c, err)
			return
		}

		c.JSON(http.StatusOK, transformer.SimpleTaskTransform(task))
	}
}

func (e *Task) ImportHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		cred, err := bind(c)
		if err != nil {
			response.Response(c, err)
			return
		}

		task, err := e.NewTask(c, model3.Import, name, cred)
		if err != nil {
			response.Response(c, err)
			return
		}

		c.JSON(http.StatusOK, transformer.SimpleTaskTransform(task))
	}
}

func (e *Task) MassUpdateHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		cred, err := bind(c)
		if err != nil {
			response.Response(c, err)
			return
		}

		task, err := e.NewTask(c, model3.MassUpdate, name, cred)
		if err != nil {
			response.Response(c, err)
			return
		}

		c.JSON(http.StatusOK, transformer.SimpleTaskTransform(task))
	}
}

func (e *Task) MassDeleteHandler(name string, bind func(c *gin.Context) ([]byte, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		cred, err := bind(c)
		if err != nil {
			response.Response(c, err)
			return
		}

		task, err := e.NewTask(c, model3.MassDelete, name, cred)
		if err != nil {
			response.Response(c, err)
			return
		}

		c.JSON(http.StatusOK, transformer.SimpleTaskTransform(task))
	}
}
