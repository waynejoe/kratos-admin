package data

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"kratos-admin/pkg/toolbox/datax"
)

type Data struct {
	G *GlobalData
}

type GlobalData struct {
	AdminDB    *gorm.DB
	RDB        *redis.Client
	LocalCache datax.Cache
}

func NewData(adminDSN string, redisOpts *redis.Options) (*Data, error) {
	adminDB, err := gorm.Open(mysql.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(redisOpts)
	pubsub := datax.NewPubsub(rdb)
	disCache := datax.NewDistributeCache(rdb)
	localCache := datax.NewLevelCache(datax.NewLocalCache(200, "mome:expire_topic:v1", pubsub), disCache)
	return &Data{G: &GlobalData{
		AdminDB:    adminDB,
		RDB:        rdb,
		LocalCache: localCache,
	}}, nil
}
