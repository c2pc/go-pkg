package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/sse"
	"github.com/c2pc/go-pkg/v2/sse/models"
	"github.com/c2pc/go-pkg/v2/task/internal/model"
	"github.com/c2pc/go-pkg/v2/task/internal/repository"
	"github.com/c2pc/go-pkg/v2/task/internal/runner"
	model3 "github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/datautil"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

const (
	TaskMessage = "task"
)

type TaskMsg struct {
	Status string `json:"status"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

var (
	ErrTaskNotFound         = apperr.New("task_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Задача не найдена", translator.EN: "Task not found"}), apperr.WithCode(code.NotFound))
	ErrTaskTypeNotFound     = apperr.New("task_type_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Тип задачи не найден", translator.EN: "Task's type not found"}), apperr.WithCode(code.NotFound))
	ErrTaskCannotStop       = apperr.New("task_cannot_stop", apperr.WithTextTranslate(translator.Translate{translator.RU: "Невозможно остановить завершенную задачу", translator.EN: "Cannot stop a completed task"}), apperr.WithCode(code.NotFound))
	ErrTaskUnableRerun      = apperr.New("task_unable_rerun", apperr.WithTextTranslate(translator.Translate{translator.RU: "Невозможно запустить незавершенную задачу", translator.EN: "Unable to rerun an unfinished task"}), apperr.WithCode(code.NotFound))
	ErrTaskFileNotFound     = apperr.New("task_file_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Файл не найден", translator.EN: "File not found"}), apperr.WithCode(code.NotFound))
	ErrTaskFileStillOngoing = apperr.New("task_file_process", apperr.WithTextTranslate(translator.Translate{translator.RU: "Задача все еще выполняется", translator.EN: "The task is still ongoing"}), apperr.WithCode(code.NotFound))
	ErrTaskTypeInvalid      = apperr.New("invalid_task_type", apperr.WithTextTranslate(translator.Translate{translator.RU: "Только export задачи могут генерировать ссылки для скачивания", translator.EN: "Only export tasks can generate download links"}), apperr.WithCode(code.Aborted))
	ErrTaskStatusInvalid    = apperr.New("invalid_task_status", apperr.WithTextTranslate(translator.Translate{translator.RU: "Задача все еще находится в исполнении", translator.EN: "The task is still in progress"}), apperr.WithCode(code.Aborted))
	ErrGenerateToken        = apperr.New("token_generation_failed", apperr.WithTextTranslate(translator.Translate{translator.RU: "Не удалось сгенерировать токен", translator.EN: "Failed to generate token"}), apperr.WithCode(code.Internal))
	ErrInvalidLink          = apperr.New("invalid_link")
)

type Queue interface {
	Run(data runner.Data)
	Stop(id int)
}

type Consumers map[string]Consumer

type Consumer interface {
	Export(ctx context.Context, data []byte) (*model3.Message, error)
	Import(ctx context.Context, data []byte) (*model3.Message, error)
	MassUpdate(ctx context.Context, data []byte) (*model3.Message, error)
	MassDelete(ctx context.Context, data []byte) (*model3.Message, error)
}

type ITaskService interface {
	Trx(db *gorm.DB) ITaskService
	List(ctx context.Context, m *model2.Meta[model.Task]) error
	GetFull(ctx context.Context, taskID *int, statuses ...string) ([]model.Task, error)
	Update(ctx context.Context, id int, input TaskUpdateInput) error
	UpdateStatus(ctx context.Context, status string, ids ...int) error
	Create(ctx context.Context, input TaskCreateInput) (*model.Task, error)
	GetById(ctx context.Context, id int) (*model.Task, error)
	Stop(ctx context.Context, id int) error
	Rerun(ctx context.Context, id int) (*model.Task, error)
	RunTasks(ctx context.Context, statuses []string, ids ...int) error
	Download(ctx context.Context, id int) (string, error)
	GenerateDownloadToken(ctx context.Context, id int) (string, error)
	ValidateDownloadToken(ctx context.Context, token string, id int) error
}

type TaskService struct {
	taskRepository repository.ITaskRepository
	services       Consumers
	queue          Queue
	tokenSecret    string
	sseService     sse.SSE
}

func NewTaskService(
	taskRepository repository.ITaskRepository,
	services Consumers,
	queue Queue,
	tokenSecret string,
	sseService sse.SSE,
) TaskService {
	return TaskService{
		taskRepository: taskRepository,
		services:       services,
		queue:          queue,
		tokenSecret:    tokenSecret,
		sseService:     sseService,
	}
}

func (s TaskService) Trx(db *gorm.DB) ITaskService {
	s.taskRepository = s.taskRepository.Trx(db)
	return s
}

func (s TaskService) List(ctx context.Context, m *model2.Meta[model.Task]) error {
	taskID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation task id is empty")
	}

	return s.taskRepository.Omit("input", "output").Paginate(ctx, m, `user_id = ?`, taskID)
}

func (s TaskService) GetFull(ctx context.Context, taskID *int, statuses ...string) ([]model.Task, error) {
	var query []string
	var args []interface{}

	if taskID != nil {
		query = append(query, "id = ?")
		args = append(args, *taskID)
	}

	if len(statuses) > 0 {
		query = append(query, "status IN (?)")
		args = append(args, statuses)
	}

	return s.taskRepository.Omit("input", "output").List(ctx, &model2.Filter{
		OrderBy: []clause.ExpressionOrderBy{{"created_at", clause.OrderByAsc}},
	}, strings.Join(query, " AND "), args...)
}

func (s TaskService) GetById(ctx context.Context, id int) (*model.Task, error) {
	task, err := s.taskRepository.Omit("input").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	task.Output, err = s.decompressData(task.Output)
	if err != nil {
		return nil, err
	}

	if task.Type == model3.Export {
		fi, err := os.Stat(task.FilePath())
		if err == nil {
			size := fi.Size()
			task.FileSize = &size
		}

	}

	return task, nil
}

func (s TaskService) Download(ctx context.Context, id int) (string, error) {
	task, err := s.taskRepository.Omit("input", "output").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return "", ErrTaskNotFound
		}
		return "", err
	}

	if task.Type != model3.Export {
		return "", ErrTaskFileNotFound
	}

	if datautil.Contain(task.Status, model.StatusPending, model.StatusRunning) {
		return "", ErrTaskFileStillOngoing
	}

	if !datautil.Contain(task.Status, model.StatusSuccess) {
		return "", ErrTaskFileNotFound
	}

	if _, err := os.Stat(task.FilePath()); err != nil {
		return "", ErrTaskFileNotFound.WithError(err)
	}

	return task.FilePath(), nil
}

func (s TaskService) Stop(ctx context.Context, id int) error {
	task, err := s.taskRepository.Omit("input", "output").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if !datautil.Contain(task.Status, model.StatusPending, model.StatusRunning) {
		return ErrTaskCannotStop
	}

	task.Status = model.StatusStopped

	if err = s.taskRepository.Update(ctx, task, []interface{}{"status"}, `id = ?`, task.ID); err != nil {
		return err
	}

	s.queue.Stop(task.ID)

	type TaskStop struct {
		Id int `json:"id"`
	}

	msg := models.Message{
		Message: TaskStop{
			Id: task.ID,
		},
	}

	if err = s.sseService.SendMessage(ctx, TaskMessage, msg); err != nil {
		return err
	}

	return nil
}

func (s TaskService) Rerun(ctx context.Context, id int) (*model.Task, error) {
	task, err := s.taskRepository.Omit("output").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	if datautil.Contain(task.Status, model.StatusPending, model.StatusRunning) {
		return nil, ErrTaskUnableRerun
	}

	task, err = s.taskRepository.Create(ctx, &model.Task{
		Name:   task.Name,
		Type:   task.Type,
		UserID: task.UserID,
		Status: model.StatusPending,
		Output: nil,
		Input:  task.Input,
	})
	if err != nil {
		return nil, err
	}

	err = s.RunTasks(ctx, []string{model.StatusPending}, task.ID)
	if err != nil {
		return nil, err
	}

	msg := models.Message{
		Message: TaskMsg{
			Status: task.Status,
			Id:     task.ID,
			Name:   task.Name,
			Type:   task.Type,
		},
	}

	if err = s.sseService.SendMessage(ctx, TaskMessage, msg); err != nil {
		return nil, err
	}

	return task, nil
}

type TaskUpdateInput struct {
	Status *string
	Output *model3.Message
}

func (s TaskService) Update(ctx context.Context, id int, input TaskUpdateInput) error {
	task, err := s.taskRepository.Omit("output").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	var selects []interface{}
	if input.Status != nil && *input.Status != "" {
		task.Status = *input.Status
		selects = append(selects, "status")
	}
	if input.Output != nil {
		if task.Type == model3.Export {
			_, err = s.writeToFile(task, input.Output.GetData())
			if err != nil {
				return err
			}
		}

		d, err := json.Marshal(input.Output)
		if err != nil {
			return err
		}

		task.Output, err = s.compressData(d)
		if err != nil {
			return err
		}

		selects = append(selects, "output")
	}

	if len(selects) > 0 {
		if err = s.taskRepository.Update(ctx, task, selects, `id = ?`, task.ID); err != nil {
			return err
		}

		msg := models.Message{
			Message: TaskMsg{
				Status: *input.Status,
				Id:     id,
				Name:   task.Name,
				Type:   task.Type,
			},
		}

		if err = s.sseService.SendMessage(ctx, TaskMessage, msg); err != nil {
			return err
		}

	}

	return nil
}

func (s TaskService) UpdateStatus(ctx context.Context, status string, ids ...int) error {
	tasks, err := s.taskRepository.Omit("output").List(ctx, &model2.Filter{}, `id IN (?)`, ids)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return ErrTaskNotFound
	}
	var ids2 []int
	for _, task := range tasks {
		ids2 = append(ids2, task.ID)
	}

	if err := s.taskRepository.Update(ctx, &model.Task{Status: status}, []interface{}{"status"}, `id IN (?)`, ids2); err != nil {
		return err
	}

	return nil
}

func (s TaskService) RunTasks(ctx context.Context, statuses []string, ids ...int) error {
	var query []string
	var args []interface{}

	if len(ids) > 0 {
		query = append(query, "id IN (?)")
		args = append(args, ids)
	}

	if len(statuses) > 0 {
		query = append(query, "status IN (?)")
		args = append(args, statuses)
	}

	tasks, err := s.taskRepository.Omit("output").List(ctx, &model2.Filter{
		OrderBy: []clause.ExpressionOrderBy{{"created_at", clause.OrderByAsc}},
	}, strings.Join(query, " AND "), args...)
	if err != nil {
		return err
	}

	var runnerData []runner.Data
	for _, task := range tasks {
		data, err := s.decompressData(task.Input)
		if err != nil {
			return err
		}

		runFunc, err := s.getRunFunc(task.Type, task.Name)
		if err != nil {
			return err
		}

		runnerData = append(runnerData, runner.Data{
			ID:       task.ID,
			ClientID: task.UserID,
			Name:     task.Name,
			Type:     task.Type,
			Data:     data,
			RunFunc:  runFunc,
		})
	}

	for _, data := range runnerData {
		s.queue.Run(data)
	}

	return nil
}

type TaskCreateInput struct {
	Name string
	Type string
	Data []byte
}

func (s TaskService) Create(ctx context.Context, input TaskCreateInput) (*model.Task, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	if _, ok := model3.Types[input.Type]; !ok {
		return nil, ErrTaskTypeNotFound
	}

	var data []byte
	if input.Data != nil {
		d, err := s.compressData(input.Data)
		if err != nil {
			return nil, err
		}
		data = d
	}

	task, err := s.taskRepository.Create(ctx, &model.Task{
		Name:   input.Name,
		Type:   input.Type,
		UserID: userID,
		Status: model.StatusPending,
		Input:  data,
	})
	if err != nil {
		return nil, err
	}

	msg := models.Message{
		Message: TaskMsg{
			Status: model.StatusPending,
			Id:     task.ID,
			Name:   task.Name,
			Type:   task.Type,
		},
	}

	if err = s.sseService.SendMessage(ctx, TaskMessage, msg); err != nil {
		return nil, err
	}

	return task, nil
}

func (s TaskService) compressData(data []byte) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}

	var b bytes.Buffer

	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (s TaskService) decompressData(data []byte) ([]byte, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	var b bytes.Buffer
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (s TaskService) getRunFunc(tp string, name string) (runner.RunFunc, error) {
	srv, ok := s.services[name]
	if !ok {
		return nil, fmt.Errorf("service not found")
	}

	switch tp {
	case model3.Export:
		return srv.Export, nil
	case model3.Import:
		return srv.Import, nil
	case model3.MassUpdate:
		return srv.MassUpdate, nil
	case model3.MassDelete:
		return srv.MassDelete, nil
	default:
		return nil, fmt.Errorf("type not found")
	}
}

func (s TaskService) writeToFile(task *model.Task, data []byte) (string, error) {
	path := task.FilePath()

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return "", err
	}

	return path, nil
}

func (s TaskService) GenerateDownloadToken(ctx context.Context, id int) (string, error) {
	task, err := s.taskRepository.Omit("input", "output").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return "", ErrTaskNotFound
		}
		return "", err
	}

	if task.Type != model3.Export {
		return "", ErrTaskTypeInvalid
	}

	if task.Status != model.StatusSuccess {
		return "", ErrTaskStatusInvalid
	}

	claims := tokenverify.BuildLinkClaims(strconv.Itoa(id), 15*time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(s.tokenSecret))
	if err != nil {
		return "", ErrGenerateToken.WithError(err)
	}

	return tokenString, nil
}

func (s TaskService) ValidateDownloadToken(ctx context.Context, tokenString string, id int) error {
	claims, err := tokenverify.GetLinkClaimFromToken(tokenString, tokenverify.Secret(s.tokenSecret))
	if err != nil {
		return err
	}

	taskID, err := strconv.Atoi(claims.Link)
	if err != nil || taskID != id {
		return ErrInvalidLink.WithErrorText("Task ID in token does not match the ID in URL")
	}

	return nil
}
