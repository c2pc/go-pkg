package logger

import (
	"context"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"testing"
	"time"
)

func TestWithOperationID(t *testing.T) {
	ctx := context.Background()

	msg := WithOperationID(ctx, "msg")

	if msg != "msg" {
		t.Error("expected msg to be 'msg', got", msg)
	}

	operationID := time.Now().UTC().Unix()
	ctx2 := context.WithValue(ctx, constant.OperationID, operationID)

	msg2 := WithOperationID(ctx2, "msg2")
	expected := fmt.Sprintf("| %d | %s", operationID, "msg2")

	if msg2 != expected {
		t.Errorf("expected msg2 to be '%s', got %s", expected, msg2)
	}

	ctx3 := context.WithValue(ctx, constant.OperationID, "badOperationID")

	msg3 := WithOperationID(ctx3, "msg3")
	if msg3 != "msg3" {
		t.Error("expected msg to be 'msg3', got", msg3)
	}
}
