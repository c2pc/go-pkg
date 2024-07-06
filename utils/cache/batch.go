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

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/datautil"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"time"
)

type BatchDeleter interface {
	//ChainExecDel method is used for chain calls and must call Clone to prevent memory pollution.
	ChainExecDel(ctx context.Context) error
	//ExecDelWithKeys method directly takes keys for deletion.
	ExecDelWithKeys(ctx context.Context, keys []string) error
	//Clone method creates a copy of the BatchDeleter to avoid modifying the original object.
	Clone() BatchDeleter
	//AddKeys method adds keys to be deleted.
	AddKeys(keys ...string)
}

// BatchDeleterRedis is a concrete implementation of the BatchDeleter interface based on Redis and RocksCache.
type BatchDeleterRedis struct {
	redisClient redis.UniversalClient
	keys        []string
	rocksClient *rockscache.Client
}

// NewBatchDeleterRedis creates a new BatchDeleterRedis instance.
func NewBatchDeleterRedis(redisClient redis.UniversalClient, options rockscache.Options) *BatchDeleterRedis {
	return &BatchDeleterRedis{
		redisClient: redisClient,
		rocksClient: rockscache.NewClient(redisClient, options),
	}
}

// ExecDelWithKeys directly takes keys for batch deletion and publishes deletion information.
func (c *BatchDeleterRedis) ExecDelWithKeys(ctx context.Context, keys []string) error {
	distinctKeys := datautil.Distinct(keys)
	return c.execDel(ctx, distinctKeys)
}

// ChainExecDel is used for chain calls for batch deletion. It must call Clone to prevent memory pollution.
func (c *BatchDeleterRedis) ChainExecDel(ctx context.Context) error {
	distinctKeys := datautil.Distinct(c.keys)
	return c.execDel(ctx, distinctKeys)
}

// execDel performs batch deletion and publishes the keys that have been deleted to update the local cache information of other nodes.
func (c *BatchDeleterRedis) execDel(ctx context.Context, keys []string) error {
	if len(keys) > 0 {
		// Batch delete keys
		err := ProcessKeysBySlot(ctx, c.redisClient, keys, func(ctx context.Context, slot int64, keys []string) error {
			return c.rocksClient.TagAsDeletedBatch2(ctx, keys)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Clone creates a copy of BatchDeleterRedis for chain calls to prevent memory pollution.
func (c *BatchDeleterRedis) Clone() BatchDeleter {
	return &BatchDeleterRedis{
		redisClient: c.redisClient,
		keys:        c.keys,
		rocksClient: c.rocksClient,
	}
}

// AddKeys adds keys to be deleted.
func (c *BatchDeleterRedis) AddKeys(keys ...string) {
	c.keys = append(c.keys, keys...)
}

// GetRocksCacheOptions returns the default configuration options for RocksCache.
func GetRocksCacheOptions() rockscache.Options {
	opts := rockscache.NewDefaultOptions()
	opts.StrongConsistency = true
	opts.RandomExpireAdjustment = 0.2

	return opts
}

func GetCache[T any](ctx context.Context, rcClient *rockscache.Client, key string, expire time.Duration, fn func(ctx context.Context) (T, error)) (T, error) {
	var t T
	var write bool
	v, err := rcClient.Fetch2(ctx, key, expire, func() (s string, err error) {
		t, err = fn(ctx)
		if err != nil {
			return "", err
		}
		bs, err := json.Marshal(t)
		if err != nil {
			return "", err
		}
		write = true

		return string(bs), nil
	})
	if err != nil {
		return t, err
	}
	if write {
		return t, nil
	}
	if v == "" {
		return t, apperr.ErrDBRecordNotFound
	}
	err = json.Unmarshal([]byte(v), &t)
	if err != nil {
		return t, errors.New(fmt.Sprintf("cache json.Unmarshal failed, key:%s, value:%s, expire:%s %V", key, v, expire, err))
	}

	return t, nil
}

func BatchGetCache[T any, K comparable](
	ctx context.Context,
	rcClient *rockscache.Client,
	expire time.Duration,
	keys []K,
	keyFn func(key K) string,
	fns func(ctx context.Context, key K) (T, error),
) ([]T, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	res := make([]T, 0, len(keys))
	for _, key := range keys {
		val, err := GetCache(ctx, rcClient, keyFn(key), expire, func(ctx context.Context) (T, error) {
			return fns(ctx, key)
		})
		if err != nil {
			if apperr.Is(err, apperr.ErrDBRecordNotFound) {
				continue
			}
			return nil, err
		}
		res = append(res, val)
	}

	return res, nil
}
