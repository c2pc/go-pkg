package runner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/c2pc/go-pkg/v2/task/internal/runner"
	"github.com/c2pc/go-pkg/v2/task/model"
)

func TestNewRunner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := runner.NewRunner(ctx)

	if r.TaskResults() == nil {
		t.Error("TaskResults channel is not initialized")
	}
}

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := runner.NewRunner(ctx)

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "test",
		Data:     nil,
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			return &model.Message{Count: 100}, nil
		},
	}

	r.Run(data)

	select {
	case result := <-r.TaskResults():
		if *result.Status != runner.StatusRunning {
			t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
		}
	case <-time.After(time.Second):
		t.Error("Task did not start running")
	}
}

func TestStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := runner.NewRunner(ctx)

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "test",
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
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
		if *result.Status != runner.StatusRunning {
			t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
		}
	case <-time.After(1 * time.Second):
		t.Error("Task was not stopped in time")
	}

	time.Sleep(500 * time.Millisecond)

	r.Stop(data.ID)

	select {
	case result := <-r.TaskResults():
		if *result.Status != runner.StatusStopped {
			t.Errorf("Expected status %s, got %s", runner.StatusStopped, *result.Status)
		}
	case <-time.After(1 * time.Second):
		t.Error("Task was not stopped in time2")
	}
}

func TestExit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	r := runner.NewRunner(ctx)

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "test",
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			time.Sleep(10 * time.Second)
			return &model.Message{Count: 100}, nil
		},
	}

	r.Run(data)

	select {
	case result := <-r.TaskResults():
		if *result.Status != runner.StatusRunning {
			t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
		}
	case <-time.After(1 * time.Second):
		t.Error("Task was not stopped in time")
	}

	select {
	case result, ok := <-r.TaskResults():
		if !ok {
			break
		}
		fmt.Println(result)
		if *result.Status != runner.StatusStopped {
			t.Errorf("Expected status %s, got %s", runner.StatusStopped, *result.Status)
		}
	case <-time.After(1 * time.Second):
		t.Error("Task was not stopped in time2")
	}
}

func TestRunFuncError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := runner.NewRunner(ctx)

	expectedErr := errors.New("task failed")

	data := runner.Data{
		ID:       1,
		ClientID: 123,
		Name:     "Failing Task",
		Type:     "test",
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			return nil, expectedErr
		},
	}

	r.Run(data)

	select {
	case result := <-r.TaskResults():
		if *result.Status != runner.StatusRunning {
			t.Errorf("Expected status %s, got %s", runner.StatusRunning, *result.Status)
		}
	case <-time.After(time.Second):
		t.Error("Task did not fail as expected")
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != runner.StatusFailed {
			t.Errorf("Expected status %s, got %s", runner.StatusFailed, *result.Status)
		}
		if !errors.Is(expectedErr, result.Error) {
			t.Errorf("Expected error %v, got %v", expectedErr, result.Error)
		}
	case <-time.After(time.Second):
		t.Error("Task did not fail as expected")
	}
}

func TestConcurrentRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errTask = errors.New("task failed")

	r := runner.NewRunner(ctx)

	numTasks := 1000
	waitDuration := time.Millisecond

	for i := 0; i < numTasks; i++ {
		func(id int) {
			data := runner.Data{
				ID:       numTasks + i,
				ClientID: numTasks,
				Name:     fmt.Sprintf("Task %d", numTasks%2),
				Type:     "test",
				RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
					time.Sleep(time.Duration(id%10) * waitDuration)
					if id%13 == 0 {
						return nil, errTask
					}
					return &model.Message{Count: 100}, nil
				},
			}
			r.Run(data)
		}(i)
	}

	go func() {
		time.Sleep(waitDuration)
		for i := 0; i < numTasks; i++ {
			func(id int) {
				data := runner.Data{
					ID:       id,
					ClientID: id % 10,
					Name:     fmt.Sprintf("Task %d", id%10),
					Type:     "test",
					RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
						time.Sleep(time.Duration(id%10) * waitDuration)
						if id%15 == 0 {
							return nil, errTask
						}
						return &model.Message{Count: 100}, nil
					},
				}
				r.Run(data)
			}(i)
		}
	}()

	go func() {
		time.Sleep(waitDuration)
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
				if *result.Status == runner.StatusRunning {
					continue
				} else if *result.Status == runner.StatusCompleted {
					completedMap[result.ID] = struct{}{}
					completedTasks++
					fmt.Printf("task-%d %s\n", result.ID, runner.StatusCompleted)
				} else if *result.Status == runner.StatusStopped {
					stoppedMap[result.ID] = struct{}{}
					completedTasks++
					fmt.Printf("task-%d %s\n", result.ID, runner.StatusStopped)
				} else if *result.Status == runner.StatusFailed {
					failedMap[result.ID] = struct{}{}
					completedTasks++
					fmt.Printf("task-%d %s\n", result.ID, runner.StatusFailed)
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

	fmt.Printf("Completed %d / Stopped %d / Failed %d\n", len(completedMap), len(stoppedMap), len(failedMap))
}
