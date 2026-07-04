package redis

import "github.com/redis/go-redis/v9"

// RedisClient returns a redis client which the application
// will use for short-lived storage tasks.
func RedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		// might add later
	})
}
