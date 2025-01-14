package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/c2pc/go-pkg/v2/task/model"
)

func TestCompleted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := NewRunner(ctx)

	// Test Run method
	data := Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "Type1",
		Data:     []byte("test data"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			time.Sleep(100 * time.Millisecond)
			return &model.Message{Count: 100}, nil
		},
	}
	data2 := Data{
		ID:       2,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "Type2",
		Data:     []byte("test data2"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			time.Sleep(100 * time.Millisecond)
			return &model.Message{Count: 100}, nil
		},
	}

	r.Run(data)
	r.Run(data2)

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusRunning {
			t.Errorf("Expected status %s, got %s", StatusRunning, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusCompleted {
			t.Errorf("Expected status %s, got %s", StatusCompleted, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusRunning {
			t.Errorf("Expected status %s, got %s", StatusRunning, *result.Status)
		}
		if result.Task.ID != data2.ID {
			t.Errorf("Expected task ID %d, got %d", data2.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusCompleted {
			t.Errorf("Expected status %s, got %s", StatusCompleted, *result.Status)
		}
		if result.Task.ID != data2.ID {
			t.Errorf("Expected task ID %d, got %d", data2.ID, result.Task.ID)
		}
	}
}

func TestStopped(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := NewRunner(ctx)

	// Test Run method
	data := Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "Type1",
		Data:     []byte("test data"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			time.Sleep(1 * time.Second)
			return &model.Message{Count: 100}, nil
		},
	}
	data2 := Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "Type1",
		Data:     []byte("test data"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			time.Sleep(1 * time.Second)
			return &model.Message{Count: 100}, nil
		},
	}

	r.Run(data)
	r.Run(data2)

	go func() {
		time.Sleep(500 * time.Millisecond)
		r.Stop(data.ID)
		time.Sleep(500 * time.Millisecond)
		r.Stop(data2.ID)
	}()

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusRunning {
			t.Errorf("Expected status %s, got %s", StatusRunning, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusStopped {
			t.Errorf("Expected status %s, got %s", StatusStopped, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}
}

func TestFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := NewRunner(ctx)

	// Test Run method
	data := Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "Type1",
		Data:     []byte("test data"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			return nil, errors.New("some error")
		},
	}
	data2 := Data{
		ID:       2,
		ClientID: 123,
		Name:     "Test Task2",
		Type:     "Type2",
		Data:     []byte("test data2"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			return nil, errors.New("some error2")
		},
	}

	r.Run(data)
	r.Run(data2)

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusRunning {
			t.Errorf("Expected status %s, got %s", StatusRunning, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusFailed {
			t.Errorf("Expected status %s, got %s", StatusFailed, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusRunning {
			t.Errorf("Expected status %s, got %s", StatusRunning, *result.Status)
		}
		if result.Task.ID != data2.ID {
			t.Errorf("Expected task ID %d, got %d", data2.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusFailed {
			t.Errorf("Expected status %s, got %s", StatusFailed, *result.Status)
		}
		if result.Task.ID != data2.ID {
			t.Errorf("Expected task ID %d, got %d", data2.ID, result.Task.ID)
		}
	}
}

func TestRunExistsTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := NewRunner(ctx)

	// Test Run method
	data := Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "Type1",
		Data:     []byte("test data"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			return &model.Message{Count: 100}, nil
		},
	}

	r.Run(data)
	r.Run(data)

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusRunning {
			t.Errorf("Expected status %s, got %s", StatusRunning, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusCompleted {
			t.Errorf("Expected status %s, got %s", StatusCompleted, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}
}

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	r := NewRunner(ctx)

	// Test Run method
	data := Data{
		ID:       1,
		ClientID: 123,
		Name:     "Test Task",
		Type:     "Type1",
		Data:     []byte("test data"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			time.Sleep(1 * time.Second)
			return &model.Message{Count: 100}, nil
		},
	}

	data2 := Data{
		ID:       2,
		ClientID: 123,
		Name:     "Test Task2",
		Type:     "Type2",
		Data:     []byte("test data"),
		RunFunc: func(ctx context.Context, data []byte) (*model.Message, error) {
			time.Sleep(1 * time.Second)
			return &model.Message{Count: 100}, nil
		},
	}

	r.Run(data)
	r.Run(data2)

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusRunning {
			t.Errorf("Expected status %s, got %s", StatusRunning, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusStopped {
			t.Errorf("Expected status %s, got %s", StatusStopped, *result.Status)
		}
		if result.Task.ID != data2.ID {
			t.Errorf("Expected task ID %d, got %d", data2.ID, result.Task.ID)
		}
	}

	select {
	case result := <-r.TaskResults():
		if *result.Status != StatusStopped {
			t.Errorf("Expected status %s, got %s", StatusStopped, *result.Status)
		}
		if result.Task.ID != data.ID {
			t.Errorf("Expected task ID %d, got %d", data.ID, result.Task.ID)
		}
	}
}
