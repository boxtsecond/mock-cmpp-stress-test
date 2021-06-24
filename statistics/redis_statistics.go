package statistics

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/config"
	"strconv"
	"strings"
	"time"
)

type RedisStatistics struct {
	cfg    *config.RedisConfig
	ctx    context.Context
	Logger *zap.Logger
	Client *redis.Client
}

func (s *RedisStatistics) Init(logger *zap.Logger, ctx context.Context) {
	s.ctx = ctx
	s.cfg = config.ConfigObj.Redis
	s.Logger = logger
}

func (s *RedisStatistics) Start() error {
	if !s.cfg.Enable {
		return nil
	}

	s.Client = s.NewRedisClient()
	if _, err := s.Client.Ping(s.Client.Context()).Result(); err != nil {
		s.Logger.Fatal("Redis Start Error", zap.Error(err))
		return err
	}

	return nil
}

func (s *RedisStatistics) Stop() error {
	return s.Client.Close()
}

func (s *RedisStatistics) NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", s.cfg.IP, s.cfg.Port),
		Password:    s.cfg.Password,
		DB:          int(s.cfg.DB),
		DialTimeout: time.Duration(s.cfg.TimeOut) * time.Second,
	})
}

func (s *RedisStatistics) SaveMachineStatistics(tickerCount int, cpu, mem, disk float64) error {
	key := "Machine"
	status := s.Client.ZAdd(s.ctx, key, &redis.Z{
		Score:  float64(tickerCount),
		Member: []float64{cpu * 100, mem * 100, disk * 100},
	})
	return status.Err()
}

func (s *RedisStatistics) SavePackerStatistics(tickerCount int) error {
	pipeline := s.Client.Pipeline()
	keys := []string{
		"Submit_Total", "Submit_Success",
		"Submit_Resp_Total", "Submit_Resp_Total",
		"Deliver_Total", "Deliver_Success",
		"Deliver_Resp_Total", "Deliver_Resp_Success",
	}
	for _, k := range keys {
		pipeline.Get(s.ctx, k)
	}
	cmd, err := pipeline.Exec(s.ctx)
	if err != nil {
		return err
	}

	members := make([]string, 0)
	for _, c := range cmd {
		result, _ := c.(*redis.IntCmd).Result()
		members = append(members, string(result))
	}

	status := s.Client.ZAdd(s.ctx, "Packer", &redis.Z{
		Score:  float64(tickerCount),
		Member: strings.Join(members, ","),
	})
	return status.Err()
}

func (s *RedisStatistics) AddPackerStatistics(name string, success bool) {
	key := []string{fmt.Sprintf("%s_Total", name)}
	if success {
		key = append(key, fmt.Sprintf("%s_Success", name))
	}
	for _, k := range key {
		if err := s.Increase(k); err != nil {
			s.Logger.Error("[Collect][AddPackerStatistics] Error",
				zap.String("Key", k),
				zap.Error(err))
		}
	}
}

func (s *RedisStatistics) Increase(key string) error {
	status := s.Client.IncrBy(s.ctx, key, 1)
	if status.Err() != nil {
		s.Logger.Error("[RedisStatistics][Increase] Error",
			zap.String("Key", key),
			zap.Error(status.Err()))
	}
	return status.Err()
}

func (s *RedisStatistics) GetXAxisStart(tickerCount int) int {
	return 0
}

func (s *RedisStatistics) GetXAxisLength(tickerCount int) int {
	return tickerCount
}

func (s *RedisStatistics) GetMachineStatistics(tickerCount int) (err error, cpu, mem, disk []float64) {
	//offset := int64(0)
	//for {
	//	vals, e := s.Client.ZRangeByScore(s.ctx, "Machine", &redis.ZRangeBy{
	//		Offset: offset,
	//		Count:  5000,
	//	}).Result()
	//
	//	if e != nil {
	//		err = e
	//		return
	//	}
	//
	//	for _, v := range vals {
	//		vStrArr := strings.Split(v, ",")
	//		if len(vStrArr) == 3 {
	//			cpu = append(cpu, vStrArr[0])
	//			mem = append(mem, vStrArr[1]+"%")
	//			disk = append(disk, vStrArr[2]+"%")
	//		}
	//	}
	//
	//	if len(vals) < 5000 {
	//		return
	//	} else {
	//		offset += 5000
	//	}
	//}
	return
}

func (s *RedisStatistics) GetPackerStatistics(tickerCount int) (error, *[][]uint64) {
	offset := int64(0)
	result := make([][]uint64, 0)
	for {
		vals, e := s.Client.ZRangeByScore(s.ctx, "Packer", &redis.ZRangeBy{
			Offset: offset,
			Count:  5000,
		}).Result()

		if e != nil {
			return e, nil
		}

		for _, v := range vals {
			vStrArr := strings.Split(v, ",")
			if len(vStrArr) == 3 {
				for i, vStr := range vStrArr {
					vInt, _ := strconv.Atoi(vStr)
					result[2*i] = append(result[2*i], uint64(vInt))
					result[2*i+1] = append(result[2*i+1], uint64(vInt))
				}
			}
		}

		if len(vals) < 5000 {
			return nil, &result
		} else {
			offset += 5000
		}
	}
}
