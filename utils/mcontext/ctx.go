// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mcontext

import (
	"context"
	"github.com/c2pc/go-pkg/v2/utils/constant"
)

func WithOpUserIDContext(ctx context.Context, opUserID int) context.Context {
	return context.WithValue(ctx, constant.OpUserID, opUserID)
}

func WithOpDeviceIDContext(ctx context.Context, device int) context.Context {
	return context.WithValue(ctx, constant.OpDeviceID, device)
}

func WithOperationIDContext(ctx context.Context, operationID string) context.Context {
	return context.WithValue(ctx, constant.OperationID, operationID)
}

func SetOperationID(ctx context.Context, operationID int) context.Context {
	return context.WithValue(ctx, constant.OperationID, operationID)
}

func SetOpUserID(ctx context.Context, opUserID int) context.Context {
	return context.WithValue(ctx, constant.OpUserID, opUserID)
}

func SetOpDeviceID(ctx context.Context, opDeviceID int) context.Context {
	return context.WithValue(ctx, constant.OpDeviceID, opDeviceID)
}

func GetOperationID(ctx context.Context) (string, bool) {
	if ctx.Value(constant.OperationID) != nil {
		s, ok := ctx.Value(constant.OperationID).(string)
		if ok {
			return s, true
		}
	}
	return "", false
}

func GetOpUserID(ctx context.Context) (int, bool) {
	if ctx.Value(constant.OpUserID) != nil {
		s, ok := ctx.Value(constant.OpUserID).(int)
		if ok {
			return s, true
		}
	}
	return 0, false
}

func GetOpDeviceID(ctx context.Context) (int, bool) {
	if ctx.Value(constant.OpDeviceID) != nil {
		s, ok := ctx.Value(constant.OpDeviceID).(int)
		if ok {
			return s, true
		}
	}
	return 0, false
}
