package runner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/c2pc/go-pkg/v2/task/internal/runner"
	"github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/level"
)

func TestNewRunner(t *testing.T) {
	fmt.Println("TestNewRunner")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := runner.NewRunner(ctx, level.PRODUCTION)

	if r.TaskResults() == nil {
		t.Error("TaskResults channel is not initialized")
	}
}

func TestRun(t *testing.T) {
	fmt.Println("TestRun")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := runner.NewRunner(ctx, level.PRODUCTION)

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "test",
		Data:     nil,
		RunFunc: func(ctx context.Context, id int, data []byte, msqChan chan<- *model.Message) (*model.Message, error) {
			for i := 0; i < 10; i++ {
				msqChan <- &model.Message{Count: i * 10}
			}
			return &model.Message{Count: 110}, nil
		},
	}

	r.Run(data)

	for {
		select {
		case result := <-r.TaskResults():
			if result.Status != nil {
				if *result.Status != runner.StatusRunning {
					t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
				}
				return
			}
		case <-time.After(time.Second):
			t.Error("Task did not start running")
		}
	}

}

func TestStop(t *testing.T) {
	fmt.Println("TestStop")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := runner.NewRunner(ctx, level.PRODUCTION)

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "test",
		RunFunc: func(ctx context.Context, id int, data []byte, msqChan chan<- *model.Message) (*model.Message, error) {
			for i := 0; i < 10; i++ {
				msqChan <- &model.Message{Count: i * 10}
			}
			select {
			case <-time.After(2 * time.Second): // Имитация долгой работы
				return &model.Message{Count: 100}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	r.Run(data)

	select {
	case result := <-r.TaskResults():
		if result.Status != nil {
			if *result.Status != runner.StatusRunning {
				t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
			}
			return
		}
	case <-time.After(1 * time.Second):
		t.Error("Task was not stopped in time")
	}

	time.Sleep(500 * time.Millisecond)

	r.Stop(data.ID)

	for {
		select {
		case result := <-r.TaskResults():
			if result.Status != nil {
				if *result.Status != runner.StatusStopped {
					t.Errorf("Expected status %s, got %s", runner.StatusStopped, *result.Status)
				}
				return
			}
		case <-time.After(1 * time.Second):
			t.Error("Task was not stopped in time2")
		}
	}

}

func TestExit(t *testing.T) {
	fmt.Println("TestExit")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	r := runner.NewRunner(ctx, level.PRODUCTION)

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "test",
		RunFunc: func(ctx context.Context, id int, data []byte, msqChan chan<- *model.Message) (*model.Message, error) {
			for i := 0; i < 10; i++ {
				msqChan <- &model.Message{Count: i * 10}
			}
			time.Sleep(10 * time.Second)
			return &model.Message{Count: 100}, nil
		},
	}

	r.Run(data)

	select {
	case result := <-r.TaskResults():
		if result.Status != nil {
			if *result.Status != runner.StatusRunning {
				t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
			}
		}

	case <-time.After(1 * time.Second):
		t.Error("Task was not stopped in time")
	}

	select {
	case result, ok := <-r.TaskResults():
		if !ok {
			break
		}
		if result.Status != nil {
			if *result.Status != runner.StatusStopped {
				t.Errorf("Expected status %s, got %s", runner.StatusStopped, *result.Status)
			}
			return
		}
	case <-time.After(1 * time.Second):
		t.Error("Task was not stopped in time2")
	}
}

func TestRunFuncError(t *testing.T) {
	fmt.Println("TestRunFuncError")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := runner.NewRunner(ctx, level.PRODUCTION)

	expectedErr := errors.New("task failed")

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Failing Task",
		Type:     "test",
		RunFunc: func(ctx context.Context, id int, data []byte, msqChan chan<- *model.Message) (*model.Message, error) {
			for i := 0; i < 10; i++ {
				msqChan <- &model.Message{Count: i * 10}
			}
			panic(expectedErr)
		},
	}

	r.Run(data)

	select {
	case result := <-r.TaskResults():
		if result.Status != nil {
			if *result.Status != runner.StatusRunning {
				t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
			}
		}
	case <-time.After(time.Second):
		t.Error("Task did not fail as expected")
	}

	for {
		select {
		case result := <-r.TaskResults():
			if result.Status != nil {
				if *result.Status != runner.StatusFailed {
					t.Errorf("Expected status %s, got %s", runner.StatusFailed, *result.Status)
				}
				if !apperr.Is(apperr.ErrInternal.WithError(expectedErr), result.Error) {
					t.Errorf("Expected error %v, got %v", apperr.ErrInternal.WithError(expectedErr), result.Error)
				}
				return
			}
		case <-time.After(time.Second):
			t.Error("Task did not fail as expected")
		}
	}

}

func TestConcurrentRun(t *testing.T) {
	fmt.Println("TestConcurrentRun")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errTask = errors.New("task failed")

	r := runner.NewRunner(ctx, level.PRODUCTION)

	numTasks := 1000

	for i := 0; i < numTasks; i++ {
		go func(id int) {
			data := runner.Data{
				ID:       i,
				ClientID: numTasks % 10,
				Name:     fmt.Sprintf("Task %d", numTasks%200),
				Type:     "test",
				RunFunc: func(ctx context.Context, id int, data []byte, msqChan chan<- *model.Message) (*model.Message, error) {
					for i := 0; i < 10; i++ {
						msqChan <- &model.Message{Count: i * 10}
					}
					time.Sleep(1 * time.Millisecond)
					if id%13 == 0 {
						panic(errTask)
					}
					return &model.Message{Count: 100}, nil
				},
			}
			r.Run(data)
		}(i)
	}

	for i := numTasks; i < numTasks*2; i++ {
		func(id int) {
			data := runner.Data{
				ID:       id,
				ClientID: id,
				Name:     fmt.Sprintf("Task %d", id),
				Type:     "test",
				RunFunc: func(ctx context.Context, id int, data []byte, msqChan chan<- *model.Message) (*model.Message, error) {
					for i := 0; i < 10; i++ {
						msqChan <- &model.Message{Count: i * 10}
					}
					time.Sleep(100 * time.Nanosecond)
					if id%15 == 0 {
						panic(errTask)
					}
					return &model.Message{Count: 100}, nil
				},
			}
			r.Run(data)
		}(i)
	}

	go func() {
		for i := numTasks - 100; i < numTasks+100; i++ {
			r.Stop(i)
		}
	}()

	completedMap := make(map[int]struct{})
	stoppedMap := make(map[int]struct{})
	failedMap := make(map[int]struct{})
	completedTasks := 0

	func() {
		for completedTasks < numTasks*2 {
			select {
			case result := <-r.TaskResults():
				if result.Status != nil {
					if *result.Status == runner.StatusRunning {
						continue
					} else if *result.Status == runner.StatusCompleted {
						completedMap[result.ID] = struct{}{}
						completedTasks++
					} else if *result.Status == runner.StatusStopped {
						stoppedMap[result.ID] = struct{}{}
						completedTasks++
					} else if *result.Status == runner.StatusFailed {
						failedMap[result.ID] = struct{}{}
						completedTasks++
					}
				}
			case <-time.After(5 * time.Second):
				t.Error("Time completed", len(completedMap)+len(stoppedMap))
				return
			}
		}
	}()

	for i := 0; i < numTasks*2; i++ {
		_, ok := completedMap[i]
		_, ok2 := stoppedMap[i]
		_, ok3 := failedMap[i]

		if !(ok || ok2 || ok3) {
			t.Errorf("Task %d did not completed/stopped", i)
		}
	}

	if len(completedMap)+len(stoppedMap)+len(failedMap) != numTasks*2 {
		t.Errorf("Expected %d completed tasks, got %d", numTasks*2, len(completedMap)+len(stoppedMap)+len(failedMap))
	}

	fmt.Printf("ALL %d / Completed %d / Stopped %d / Failed %d\n", numTasks*2, len(completedMap), len(stoppedMap), len(failedMap))
}
