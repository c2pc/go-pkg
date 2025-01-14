package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/constant"
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
	taskResults   chan TaskResult
	runnerMU      sync.Mutex
	runner        chan Data
	stopper       chan int
	activeTasksMU sync.Mutex
	activeTasks   map[int]chan struct{}
	ctx           context.Context
}

func NewRunner(ctx context.Context) *Runner {
	runner := &Runner{
		taskResults:   make(chan TaskResult),
		runnerMU:      sync.Mutex{},
		runner:        make(chan Data),
		stopper:       make(chan int),
		activeTasksMU: sync.Mutex{},
		activeTasks:   make(map[int]chan struct{}),
		ctx:           ctx,
	}
	go runner.listen()
	return runner
}

func (r *Runner) TaskResults() chan TaskResult {
	return r.taskResults
}

func (r *Runner) Run(data Data) {
	if r.checkActiveTask(data.ID) {
		return
	}
	r.setActiveTasks(data.ID)
	r.runner <- data
}

func (r *Runner) Stop(id int) {
	r.stopper <- id
}

func (r *Runner) listen() {
	for {
		select {
		case data := <-r.runner:
			go r.run(data)
		case id := <-r.stopper:
			go r.stop(id)
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Runner) run(data Data) {
	r.runnerMU.Lock()
	defer r.runnerMU.Unlock()

	ch, exists := r.getActiveTask(data.ID)
	if !exists {
		return
	}

	task := Task{ID: data.ID, ClientID: data.ClientID, Name: data.Name, Type: data.Type, RanAt: time.Now()}

	if r.ctx.Err() != nil {
		status := StatusStopped
		r.taskResults <- TaskResult{Task: task, Status: &status}
		return
	}

	ctx, cancel := context.WithCancel(context.WithValue(r.ctx, constant.OperationID, fmt.Sprintf("runner-task-%d", task.ID)))
	defer cancel()

	status := StatusRunning
	r.taskResults <- TaskResult{Task: task, Status: &status}

	done := make(chan struct{})

	go func(ctx context.Context) {
		msg, err := data.RunFunc(ctx, data.Data)
		task.EndedAt = time.Now()
		if err != nil {
			status := StatusFailed
			r.taskResults <- TaskResult{Task: task, Status: &status, Error: err}
		} else if ctx.Err() != nil {
			status := StatusStopped
			r.taskResults <- TaskResult{Task: task, Status: &status, Message: msg}
		} else {
			status := StatusCompleted
			r.taskResults <- TaskResult{Task: task, Status: &status, Message: msg}
		}
		r.deleteActiveTasks(task.ID)
		close(done)
	}(ctx)

	select {
	case <-ctx.Done():
		return
	case <-ch:
		cancel()
		return
	case <-done:
		return
	}
}

func (r *Runner) stop(id int) {
	if ch, exists := r.getActiveTask(id); exists {
		ch <- struct{}{}
	}
}

func (r *Runner) deleteActiveTasks(ids ...int) {
	r.activeTasksMU.Lock()
	defer r.activeTasksMU.Unlock()
	for _, id := range ids {
		delete(r.activeTasks, id)
	}
}

func (r *Runner) setActiveTasks(ids ...int) {
	r.activeTasksMU.Lock()
	defer r.activeTasksMU.Unlock()
	for _, id := range ids {
		r.activeTasks[id] = make(chan struct{})
	}
}

func (r *Runner) checkActiveTask(id int) bool {
	r.activeTasksMU.Lock()
	defer r.activeTasksMU.Unlock()
	_, exists := r.activeTasks[id]
	return exists
}

func (r *Runner) getActiveTask(id int) (chan struct{}, bool) {
	r.activeTasksMU.Lock()
	defer r.activeTasksMU.Unlock()
	ch, exists := r.activeTasks[id]
	return ch, exists
}
