package gokvstores

import (
	"net"
	"time"

	conv "github.com/cstockton/go-conv"
	redis "gopkg.in/redis.v5"
)

// ----------------------------------------------------------------------------
// Client
// ----------------------------------------------------------------------------

// RedisClient is an interface thats allows to use Redis cluster or a redis single client seamlessly.
type RedisClient interface {
	Ping() *redis.StatusCmd
	Exists(key string) *redis.BoolCmd
	Del(keys ...string) *redis.IntCmd
	FlushDb() *redis.StatusCmd
	Close() error
	Process(cmd redis.Cmder) error
	Get(key string) *redis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	HGetAll(key string) *redis.StringStringMapCmd
	HMSet(key string, fields map[string]string) *redis.StatusCmd
	SMembers(key string) *redis.StringSliceCmd
	SAdd(key string, members ...interface{}) *redis.IntCmd
}

// RedisClientOptions are Redis client options.
type RedisClientOptions struct {
	Network            string
	Addr               string
	Dialer             func() (net.Conn, error)
	Password           string
	DB                 int
	MaxRetries         int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	ReadOnly           bool
}

// RedisClusterOptions are Redis cluster options.
type RedisClusterOptions struct {
	Addrs              []string
	MaxRedirects       int
	ReadOnly           bool
	RouteByLatency     bool
	Password           string
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
}

// ----------------------------------------------------------------------------
// Store
// ----------------------------------------------------------------------------

// RedisStore is the Redis implementation of KVStore.
type RedisStore struct {
	client     RedisClient
	expiration time.Duration
}

// Get returns value for the given key.
func (r *RedisStore) Get(key string) (interface{}, error) {
	cmd := redis.NewCmd("get", key)

	if err := r.client.Process(cmd); err != nil {
		return nil, err
	}

	return cmd.Val(), cmd.Err()
}

// Set sets the value for the given key.
func (r *RedisStore) Set(key string, value interface{}) error {
	return r.client.Set(key, value, r.expiration).Err()
}

// GetMap returns map for the given key.
func (r *RedisStore) GetMap(key string) (map[string]interface{}, error) {
	values, err := r.client.HGetAll(key).Result()
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, nil
	}

	newValues := make(map[string]interface{}, len(values))
	for k, v := range values {
		newValues[k] = v
	}

	return newValues, nil
}

// SetMap sets map for the given key.
func (r *RedisStore) SetMap(key string, values map[string]interface{}) error {
	newValues := make(map[string]string, len(values))

	for k, v := range values {
		newValues[k] = conv.String(v)
	}

	return r.client.HMSet(key, newValues).Err()
}

// GetSlice returns slice for the given key.
func (r *RedisStore) GetSlice(key string) ([]interface{}, error) {
	values, err := r.client.SMembers(key).Result()
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, nil
	}

	newValues := make([]interface{}, len(values))
	for _, v := range values {
		newValues = append(newValues, v)
	}

	return newValues, nil
}

// SetSlice sets map for the given key.
func (r *RedisStore) SetSlice(key string, values []interface{}) error {
	for _, v := range values {
		if v != nil {
			if err := r.client.SAdd(key, v).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

// AppendSlice appends values to the given slice.
func (r *RedisStore) AppendSlice(key string, values ...interface{}) error {
	return r.SetSlice(key, values)
}

// Exists checks key existence.
func (r *RedisStore) Exists(key string) (bool, error) {
	cmd := r.client.Exists(key)
	return cmd.Val(), cmd.Err()
}

// Delete deletes key.
func (r *RedisStore) Delete(key string) error {
	return r.client.Del(key).Err()
}

// Flush flushes the current database.
func (r *RedisStore) Flush() error {
	return r.client.FlushDb().Err()
}

// Close closes the client connection.
func (r *RedisStore) Close() error {
	return r.client.Close()
}

// NewRedisClientStore returns Redis client instance of KVStore.
func NewRedisClientStore(options *RedisClientOptions, expiration time.Duration) (KVStore, error) {
	opts := &redis.Options{
		Network:            options.Network,
		Addr:               options.Addr,
		Dialer:             options.Dialer,
		Password:           options.Password,
		DB:                 options.DB,
		MaxRetries:         options.MaxRetries,
		DialTimeout:        options.DialTimeout,
		ReadTimeout:        options.ReadTimeout,
		WriteTimeout:       options.WriteTimeout,
		PoolSize:           options.PoolSize,
		PoolTimeout:        options.PoolTimeout,
		IdleTimeout:        options.IdleTimeout,
		IdleCheckFrequency: options.IdleCheckFrequency,
		ReadOnly:           options.ReadOnly,
	}

	client := redis.NewClient(opts)

	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	return &RedisStore{
		client:     client,
		expiration: expiration,
	}, nil
}

// NewRedisClusterStore returns Redis cluster client instance of KVStore.
func NewRedisClusterStore(options *RedisClusterOptions, expiration time.Duration) (KVStore, error) {
	opts := &redis.ClusterOptions{
		Addrs:              options.Addrs,
		MaxRedirects:       options.MaxRedirects,
		ReadOnly:           options.ReadOnly,
		RouteByLatency:     options.RouteByLatency,
		Password:           options.Password,
		DialTimeout:        options.DialTimeout,
		ReadTimeout:        options.ReadTimeout,
		WriteTimeout:       options.WriteTimeout,
		PoolSize:           options.PoolSize,
		PoolTimeout:        options.PoolTimeout,
		IdleTimeout:        options.IdleTimeout,
		IdleCheckFrequency: options.IdleCheckFrequency,
	}

	client := redis.NewClusterClient(opts)

	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	return &RedisStore{
		client:     client,
		expiration: expiration,
	}, nil
}
