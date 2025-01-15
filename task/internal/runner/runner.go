package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/c2pc/go-pkg/v2/task/internal/logger"
	"github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/level"
)

const (
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusStopped   = "stopped"
)

type Task struct {
	ID       int
	ClientID int
	Name     string
	Type     string
	RanAt    time.Time
	EndedAt  time.Time
}

type TaskResult struct {
	Task
	Status  *string
	Message *model.Message
	Error   error
}

type RunFunc func(ctx context.Context, data []byte) (*model.Message, error)

type Data struct {
	ID       int
	ClientID int
	Name     string
	Type     string
	Data     []byte
	RunFunc  RunFunc
}

type Runner struct {
	taskResults  chan TaskResult
	runner       chan Data
	stopper      chan int
	activeTasks  sync.Map
	clientQueues sync.Map
	nameLocks    sync.Map // Prevent simultaneous tasks with the same Name
	semaphore    chan struct{}
	ctx          context.Context
	debug        string
}

func NewRunner(ctx context.Context, debug string) *Runner {
	runner := &Runner{
		taskResults: make(chan TaskResult, 100), // Buffered channel for task results
		runner:      make(chan Data, 50),        // Buffered channel for data
		stopper:     make(chan int, 50),         // Buffered channel for stopper
		semaphore:   make(chan struct{}, 15),    // Limit concurrency to 15 tasks
		ctx:         ctx,
		debug:       debug,
	}

	runner.printf(runner.ctx, "Runner initialized")

	go runner.listen()

	return runner
}

func (r *Runner) TaskResults() chan TaskResult {
	return r.taskResults
}

func (r *Runner) Run(data Data) {
	r.printf(r.ctx, "Received task: ID=%d, ClientID=%d, Name=%s", data.ID, data.ClientID, data.Name)
	clientQueue := r.getClientQueue(data.ClientID)

	// Добавляем задачу в очередь клиента
	clientQueue <- data

	// Запускаем обработчик очереди, если он еще не запущен
	if len(clientQueue) == 1 {
		r.startClientQueueProcessor(data.ClientID)
	}
}

func (r *Runner) Stop(id int) {
	defer func() {
		recover()
	}()
	r.stopper <- id
}

func (r *Runner) Shutdown() {
	r.printf(r.ctx, "Shutting Down runner")
	close(r.runner)
	close(r.stopper)
	close(r.taskResults)
	close(r.semaphore)
}

func (r *Runner) listen() {
	r.printf(r.ctx, "Runner started listening")
	for {
		select {
		case data := <-r.runner:
			go r.run(data)
		case id := <-r.stopper:
			go r.stop(id)
		case <-r.ctx.Done():
			r.Shutdown()
			return
		}
	}
}

func (r *Runner) run(data Data) {
	defer func() {
		if rec := recover(); rec != nil {
			r.printf(r.ctx, "Recovered from panic in task ID=%d: %v", data.ID, rec)
			status := StatusFailed
			r.taskResults <- TaskResult{Task: Task{ID: data.ID, ClientID: data.ClientID}, Status: &status, Error: fmt.Errorf("panic: %v", rec)}
			r.deleteActiveTask(data.ID)
			r.unlockName(data.Name)
		}
	}()

	r.setActiveTask(data.ID)
	defer r.deleteActiveTask(data.ID)

	r.printf(r.ctx, "Waiting for lock on Name=%s", data.Name)
	r.lockName(data.Name)
	defer r.unlockName(data.Name)

	r.semaphore <- struct{}{}
	defer func() { <-r.semaphore }()

	task := Task{
		ID:       data.ID,
		ClientID: data.ClientID,
		Name:     data.Name,
		Type:     data.Type,
		RanAt:    time.Now(),
	}

	if _, ok := r.getActiveTask(data.ID); !ok {
		status := StatusStopped
		r.printf(r.ctx, "Context canceled before task start: ID=%d", data.ID)
		r.sendTaskResult(TaskResult{Task: task, Status: &status})
		r.deleteActiveTask(data.ID)
		return
	}

	if r.ctx.Err() != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.WithValue(r.ctx, constant.OperationID, fmt.Sprintf("runner-task-%d", task.ID)))
	defer cancel()

	status := StatusRunning
	r.sendTaskResult(TaskResult{Task: task, Status: &status})
	r.printf(ctx, "Running task: ID=%d, ClientID=%d, Name=%s", data.ID, data.ClientID, data.Name)

	done := make(chan struct{})

	go func() {
		defer close(done)
		msg, err := data.RunFunc(ctx, data.Data)
		task.EndedAt = time.Now()
		if r.ctx.Err() != nil {
			r.printf(ctx, "Task stopped globally: ID=%d", task.ID)
		} else if ctx.Err() != nil {
			status := StatusStopped
			r.printf(ctx, "Task stopped: ID=%d", task.ID)
			r.sendTaskResult(TaskResult{Task: task, Status: &status, Message: msg})
		} else if err != nil {
			status := StatusFailed
			r.printf(ctx, "Task failed: ID=%d, Error=%v", task.ID, err)
			r.sendTaskResult(TaskResult{Task: task, Status: &status, Error: err})
		} else {
			status := StatusCompleted
			r.printf(ctx, "Task completed: ID=%d", task.ID)
			r.sendTaskResult(TaskResult{Task: task, Status: &status, Message: msg})
		}
	}()

	select {
	case <-ctx.Done():
		r.printf(ctx, "Context canceled for task ID=%d", data.ID)
		return
	case _, ok := <-r.getActiveTaskChannel(data.ID):
		if ok {
			r.printf(ctx, "Task stopped manually: ID=%d", data.ID)
			cancel()
		}
		return
	case <-done:
		r.printf(ctx, "Task completed: ID=%d", data.ID)
		return
	}

}

func (r *Runner) stop(id int) {
	r.deleteActiveTask(id)
}

func (r *Runner) setActiveTask(id int) {
	if _, exists := r.getActiveTask(id); !exists {
		r.printf(r.ctx, "runner-task-%d Setting active task: ID=%d", id, id)
		r.activeTasks.Store(id, make(chan struct{}))
	}
}

func (r *Runner) deleteActiveTask(id int) {
	if ch, exists := r.getActiveTask(id); exists {
		r.printf(r.ctx, "runner-task-%d Deleting active task: ID=%d", id, id)
		select {
		case <-ch: // Проверяем, что канал еще активен
		default:
			close(ch) // Закрываем только если он не был закрыт ранее
		}
	}
	r.activeTasks.Delete(id)
}

func (r *Runner) getActiveTask(id int) (chan struct{}, bool) {
	val, exists := r.activeTasks.Load(id)
	if !exists {
		return nil, false
	}
	return val.(chan struct{}), true
}

func (r *Runner) getActiveTaskChannel(id int) chan struct{} {
	ch, _ := r.getActiveTask(id)
	return ch
}

func (r *Runner) getClientQueue(clientID int) chan Data {
	queue, _ := r.clientQueues.LoadOrStore(clientID, make(chan Data, 100))
	return queue.(chan Data)
}

func (r *Runner) startClientQueueProcessor(clientID int) {
	queue := r.getClientQueue(clientID)
	defer r.clientQueues.Delete(clientID)

	r.printf(r.ctx, "Starting client queue processor: ClientID=%d", clientID)
	defer r.printf(r.ctx, "Client queue processor stopped: ClientID=%d", clientID)

	go func() {
		for data := range queue {
			func() {
				defer func() {
					recover()
				}()
				r.runner <- data
			}()
			if len(queue) == 0 {
				break
			}
		}
	}()
}

func (r *Runner) lockName(name string) {
	ch, loaded := r.nameLocks.LoadOrStore(name, make(chan struct{}))
	if loaded {
		r.printf(r.ctx, "Task with Name=%s is waiting for the lock", name)
		<-ch.(chan struct{}) // Ожидаем завершения предыдущей задачи
	}
	r.printf(r.ctx, "Task with Name=%s acquired the lock", name)
}

func (r *Runner) unlockName(name string) {
	if ch, exists := r.nameLocks.Load(name); exists {
		select {
		case <-ch.(chan struct{}): // Проверяем, был ли канал уже закрыт
		default:
			close(ch.(chan struct{})) // Закрываем канал только если он активен
		}
	}
}

func (r *Runner) sendTaskResult(data TaskResult) {
	defer func() {
		recover()
	}()
	r.taskResults <- data
}

func (r *Runner) printf(ctx context.Context, format string, v ...any) {
	if level.Is(r.debug, level.TEST) {
		logger.LogInfo(ctx, format, v...)
	}
}
