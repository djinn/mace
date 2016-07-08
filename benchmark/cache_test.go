package cachebench

import (
	"fmt"
	"testing"
	"time"

	"github.com/djinn/mace"
	"github.com/rif/cache2go"
	"gopkg.in/redis.v3"
)

func BenchmarkRedis(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		err := client.Set(key, key, 0).Err()
		if err != nil {
			panic(err)
		}

	}

}

func BenchmarkCache2Go(b *testing.B) {
	cache := cache2go.Cache("bench")
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		cache.Cache(key, 0*time.Second, &key)
	}
}

func BenchmarkMace(b *testing.B) {
	cache := mace.Mace("bench")
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		cache.Cache(key, 0*time.Second, &key)
	}
}
