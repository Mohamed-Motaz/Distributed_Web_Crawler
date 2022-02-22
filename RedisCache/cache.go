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
//re-add data to cache to increase its ttl
func (cache *Cache) GetCachedJobIfPresent(key string, depth int, ttl time.Duration) *CachedObj{
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

	//update cache to increase ttl
	//only if the result in cache was larger
	cache.Add(key, string(data), ttl)
	return oldJob
}

//add result to cache if no result
//with the same url and a higher depth exists
func (cache *Cache) AddIfNoLargerResultPresent(key string, value *CachedObj, ttl time.Duration){
	cached := cache.GetCachedJobIfPresent(key, 0, ttl)
	if cached == nil{
		//add to cache
		res, err := json.Marshal(value)
		if err != nil{
			return
		}
		cache.Add(key, string(res), ttl)
		return
	}

	//indeed present in cache
	//now need to decide whether I need to overwrite the value written in cache
	if value.DepthToCrawl > cached.DepthToCrawl{
		res, err := json.Marshal(value)
		if err != nil{
			return
		}
		cache.Add(key, string(res), ttl)
	}
	//else
		//depth in cache is larger
		//nothing needs to be done
		//since GetCachedJobIfPresent has already updated the ttl	
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