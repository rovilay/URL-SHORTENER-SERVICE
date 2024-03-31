package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	Client *redis.Client
}

func urlRedisKey(hash string) string {
	return fmt.Sprintf("URL:%s", hash)
}

func (r *RedisRepo) Insert(ctx context.Context, url URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to encode url: %w", err)
	}

	expiry_seconds := time.Hour * 24
	key := urlRedisKey(url.ShortHash)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	// // add url sets to manage url creation url/sequence
	// if err := txn.SAdd(ctx, "urls", key).Err(); err != nil {
	// 	txn.Discard()
	// 	return fmt.Errorf("failed to add to urls set: %w", err)
	// }

	// set expiry time
	if err := txn.Expire(ctx, key, expiry_seconds).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set expiry: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

var ErrNotExist = errors.New("url does not exist")

func (r *RedisRepo) FindByHash(ctx context.Context, hash string) (URL, error) {
	key := urlRedisKey(hash)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return URL{}, ErrNotExist
	} else if err != nil {
		return URL{}, fmt.Errorf("failed to get url: %w", err)
	}

	var url URL
	err = json.Unmarshal([]byte(value), &url)
	if err != nil {
		return URL{}, fmt.Errorf("failed to decode url json: %w", err)
	}

	return url, nil
}

type Metric int

const (
	Redirect Metric = iota
	Shorten
)

// Increment metric counter
func (r *RedisRepo) CountMetric(ctx context.Context, m Metric, hash string) error {
	key := "metrics:shorten_count"

	if m == Redirect {
		key = "metrics:redirect_count:" + hash
	}

	if err := r.Client.Incr(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to increment metric: %w", err)
	}

	return nil
}

func (r *RedisRepo) GetCountMetric(ctx context.Context, m Metric, hash string) (int, error) {
	key := "metrics:shorten_count"

	if m == Redirect {
		key = "metrics:redirect_count:" + hash
	}

	count, err := r.Client.Get(ctx, key).Result()

	if err != nil {
		return 0, fmt.Errorf("failed to get metric: %w", err)
	}

	return strconv.Atoi(count)
}

// func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
// 	key := urlIDKey(id)

// 	txn := r.Client.TxPipeline()

// 	err := txn.Del(ctx, key).Err()
// 	if errors.Is(err, redis.Nil) {
// 		txn.Discard()
// 		return ErrNotExist
// 	} else if err != nil {
// 		txn.Discard()
// 		return fmt.Errorf("delete url: %w", err)
// 	}

// 	// remove url  from sets to manage url sequence
// 	if err := txn.SRem(ctx, "urls", key).Err(); err != nil {
// 		txn.Discard()
// 		return fmt.Errorf("failed to remove from urls set: %w", err)
// 	}

// 	if _, err := txn.Exec(ctx); err != nil {
// 		return fmt.Errorf("failed to exec: %w", err)
// 	}

// 	return nil
// }

// func (r *RedisRepo) Update(ctx context.Context, url URL) error {
// 	data, err := json.Marshal(url)
// 	if err != nil {
// 		return fmt.Errorf("failed to encode url: %w", err)
// 	}

// 	key := urlIDKey(uint64(url.ID))

// 	res := r.Client.SetXX(ctx, key, string(data), 0)
// 	if err := res.Err(); err != nil {
// 		return fmt.Errorf("failed to set: %w", err)
// 	}

// 	return nil
// }

// type FindAllPage struct {
// 	Size   uint64
// 	Offset uint64
// }

// type FindResult struct {
// 	urls []URL
// 	Cursor   uint64
// }

// func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
// 	res := r.Client.SScan(ctx, "urls", page.Offset, "*", int64(page.Size))

// 	keys, cursor, err := res.Result()
// 	if err != nil {
// 		return FindResult{}, fmt.Errorf("failed to get url ids: %w", err)
// 	}

// 	if len(keys) == 0 {
// 		return FindResult{
// 			urls: []URL{},
// 		}, nil
// 	}

// 	xs, err := r.Client.MGet(ctx, keys...).Result()
// 	if err != nil {
// 		return FindResult{}, fmt.Errorf("failed to get url ids: %w", err)
// 	}

// 	urls := make([]URL, len(xs))

// 	for i, x := range xs {
// 		x := x.(string)
// 		var url URL

// 		err := json.Unmarshal([]byte(x), &url)
// 		if err != nil {
// 			return FindResult{}, fmt.Errorf("failed to decode url json: %w", err)
// 		}

// 		urls[i] = url
// 	}

// 	return FindResult{
// 		urls: urls,
// 		Cursor:   cursor,
// 	}, nil
// }
