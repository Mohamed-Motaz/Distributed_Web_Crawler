package RedisCache

import (
	logger "Distributed_Web_Crawler/Logger"
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

//a wrapper for redis

type Cache struct{
	client *redis.Client							//caching layer
	ctx context.Context						//context for redis
}

type CachedObj struct{
	UrlToCrawl string		`json:"urlToCrawl"`
	DepthToCrawl int  		`json:"depthToCrawl"`
	Results [][]string		`json:"results"`
}

// func main(){
// 	cache, _ := New("localhost", "6379")
// 	cache.Add("HI", "ITS ME MARIO!", time.Second)
// 	time.Sleep(time.Millisecond * 500)
// 	val, err := cache.Get("HI")
// 	fmt.Printf("%v, %v\n", val, err == redis.Nil) 
// }

func New(cacheHost string, cachePort string) (*Cache) {
	cache := &Cache{
		client: redis.NewClient(&redis.Options{
			Addr: cacheHost + ":" + cachePort,
			Password: "",
			DB: 0,									//default db
		}),
		ctx: context.Background(),
	}

	_, err := cache.client.Ping(cache.ctx).Result()
	if err != nil{
		logger.LogError(logger.SERVER, logger.ESSENTIAL, "Unable to connect to caching layer with error %v", err)
	}else{
		logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Successfully connected to caching layer")
	}
	return cache
}

//return result if present
//else return nil
func (cache *Cache) GetValueIfPresent(key string, depth int) [][]string{
	data, err := cache.GetBytes(key)
	if err != nil{
		return nil
	}
	oldJob := &CachedObj{}
	err = json.Unmarshal(data, oldJob)
	if err != nil{
		return nil
	}
	if oldJob.DepthToCrawl < depth{
		//not enough data present
		return nil
	}
	return oldJob.Results[:depth]
}

func (cache *Cache) Add(key string, value string, ttl time.Duration) error {
	return cache.client.Set(cache.ctx, key, value, ttl).Err()
}

func (cache *Cache) Get(key string) (string, error) {
	return cache.client.Get(cache.ctx, key).Result()
}

func (cache *Cache) GetBytes(key string) ([]byte, error) {
	return cache.client.Get(cache.ctx, key).Bytes()
}