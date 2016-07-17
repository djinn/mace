package cachebench

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/djinn/mace"
	"github.com/muesli/cache2go"
	"gopkg.in/redis.v3"
)

var (
	logger = log.New(os.Stdout, "Mace:", log.LstdFlags)
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
		err = client.Del(key).Err()
		if err != nil {
			panic(err)
		}
	}

}

func BenchmarkRedisWithExpiry(b *testing.B) {
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
		err := client.Set(key, key, 300*time.Millisecond).Err()
		if err != nil {
			panic(err)
		}
	}

}

func BenchmarkCache2Go(b *testing.B) {
	cache := cache2go.Cache("bench")
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		cache.Add(key, 0*time.Second, &key)
		cache.Delete(key)
	}
}

func BenchmarkCache2GoWithExpiry(b *testing.B) {
	cache := cache2go.Cache("benchExpiry")
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		cache.Add(key, 300*time.Second, &key)
	}
}

func BenchmarkMace(b *testing.B) {
	cache := mace.Mace("bench")
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		cache.Set(key, &key, 0*time.Millisecond)
		cache.Delete(key)
	}
}

func BenchmarkMaceWithExpiry(b *testing.B) {
	cache := mace.Mace("benchExpiry")
	//cache.SetLogger(logger)
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("k%d", i)
		cache.Set(key, &key, 300*time.Millisecond)
	}
}

func BenchmarkStringAlloc( b *testing.B) {
	
}

