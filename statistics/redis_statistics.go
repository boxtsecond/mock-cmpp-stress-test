package statistics

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/utils/log"
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
	if _, err := s.Client.Ping(s.ctx).Result(); err != nil {
		s.Logger.Fatal("Redis Start Error", zap.Error(err))
		return err
	}

	if s.cfg.ClearKey {
		s.ClearKeys()
		err := s.InitKeys()
		if err != nil {
			s.Logger.Fatal("Redis InitKeys Error", zap.Error(err))
			return err
		}
	} else {
		s.CheckKeys()
	}
	return nil
}

func (s *RedisStatistics) ClearKeys() {
	keys := make([]string, 0)
	if config.ConfigObj.ClientConfig.Enable {
		keys = []string{
			"ClientMachine", "ClientPacker",
			"Client_Submit_Total", "Client_Submit_Success",
			"Client_SubmitResp_Total", "Client_SubmitResp_Success",
			"Client_Deliver_Total", "Client_Deliver_Success",
			"Client_DeliverResp_Total", "Client_DeliverResp_Success",
		}
	}

	if config.ConfigObj.ServerConfig.Enable {
		keys = append(keys, []string{
			"ServerMachine", "ServerPacker",
			"Server_Submit_Total", "Server_Submit_Success",
			"Server_SubmitResp_Total", "Server_SubmitResp_Success",
			"Server_Deliver_Total", "Server_Deliver_Success",
			"Server_DeliverResp_Total", "Server_DeliverResp_Success",
		}...)
	}
	s.Client.Del(s.ctx, keys...)
}

func (s *RedisStatistics) InitKeys() error {
	keys := make([]string, 0)
	if config.ConfigObj.ClientConfig.Enable {
		keys = []string{
			"Client_Submit_Total", "Client_Submit_Success",
			"Client_SubmitResp_Total", "Client_SubmitResp_Success",
			"Client_Deliver_Total", "Client_Deliver_Success",
			"Client_DeliverResp_Total", "Client_DeliverResp_Success",
		}
	}

	if config.ConfigObj.ServerConfig.Enable {
		keys = append(keys, []string{
			"Server_Submit_Total", "Server_Submit_Success",
			"Server_SubmitResp_Total", "Server_SubmitResp_Success",
			"Server_Deliver_Total", "Server_Deliver_Success",
			"Server_DeliverResp_Total", "Server_DeliverResp_Success",
		}...)
	}
	pipeline := s.Client.Pipeline()
	for _, k := range keys {
		pipeline.Set(s.ctx, k, 0, 0)
	}
	_, err := pipeline.Exec(s.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *RedisStatistics) CheckKeys() {
	keys := make([]string, 0)
	if config.ConfigObj.ClientConfig.Enable {
		keys = []string{
			"Client_Submit_Total", "Client_Submit_Success",
			"Client_SubmitResp_Total", "Client_SubmitResp_Success",
			"Client_Deliver_Total", "Client_Deliver_Success",
			"Client_DeliverResp_Total", "Client_DeliverResp_Success",
		}
	}

	if config.ConfigObj.ServerConfig.Enable {
		keys = append(keys, []string{
			"Server_Submit_Total", "Server_Submit_Success",
			"Server_SubmitResp_Total", "Server_SubmitResp_Success",
			"Server_Deliver_Total", "Server_Deliver_Success",
			"Server_DeliverResp_Total", "Server_DeliverResp_Success",
		}...)
	}
	pipeline := s.Client.Pipeline()
	for _, k := range keys {
		pipeline.Get(s.ctx, k)
	}

	initKeys := make([]string, 0)
	cmds, _ := pipeline.Exec(s.ctx)
	for i, _ := range cmds {
		_, e := cmds[i].(*redis.StringCmd).Result()
		if e == redis.Nil {
			initKeys = append(initKeys, keys[i])
		}
	}
	p := s.Client.Pipeline()
	for _, k := range keys {
		p.Set(s.ctx, k, 0, 0)
	}
	_, err := p.Exec(s.ctx)
	log.Logger.Error("CheckKeys Error", zap.Error(err))

	return
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
	key := ""
	if config.ConfigObj.ClientConfig.Enable {
		key = "ClientMachine"
	}

	if config.ConfigObj.ServerConfig.Enable {
		key = "ServerMachine"
	}
	status := s.Client.ZAdd(s.ctx, key, &redis.Z{
		Score: float64(tickerCount),
		Member: strings.Join([]string{
			fmt.Sprintf("%.3f", cpu),
			fmt.Sprintf("%.3f", mem),
			fmt.Sprintf("%.3f", disk),
		}, ","),
	})
	return status.Err()
}

func (s *RedisStatistics) SavePackerStatistics(tickerCount int) error {
	keys := make([]string, 0)
	packerKey := ""
	enableCount := 0

	if config.ConfigObj.ClientConfig.Enable {
		packerKey = "ClientPacker"
		keys = []string{
			"Client_Submit_Total", "Client_Submit_Success",
			"Client_SubmitResp_Total", "Client_SubmitResp_Success",
			"Client_Deliver_Total", "Client_Deliver_Success",
			"Client_DeliverResp_Total", "Client_DeliverResp_Success",
		}
		enableCount += 1
	}

	if config.ConfigObj.ServerConfig.Enable {
		packerKey = "ServerPacker"
		keys = append(keys, []string{
			"Server_Submit_Total", "Server_Submit_Success",
			"Server_SubmitResp_Total", "Server_SubmitResp_Success",
			"Server_Deliver_Total", "Server_Deliver_Success",
			"Server_DeliverResp_Total", "Server_DeliverResp_Success",
		}...)
		enableCount += 1
	}

	if enableCount == 2 {
		packerKey = "Packer"
	}

	pipeline := s.Client.Pipeline()
	for _, k := range keys {
		pipeline.Get(s.ctx, k)
	}
	cmd, err := pipeline.Exec(s.ctx)
	if err != nil {
		return err
	}

	members := make([]string, 0)
	for _, c := range cmd {
		result, _ := c.(*redis.StringCmd).Result()
		members = append(members, result)
	}

	members = append(members, strconv.Itoa(tickerCount))

	status := s.Client.ZAdd(s.ctx, packerKey, &redis.Z{
		Score:  float64(tickerCount),
		Member: strings.Join(members, ","),
	})
	return status.Err()
}

func (s *RedisStatistics) AddPackerStatistics(source, name string, success bool) {
	key := []string{fmt.Sprintf("%s_%s_Total", source, name)}
	if success {
		key = append(key, fmt.Sprintf("%s_%s_Success", source, name))
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
	offset := 0
	interval := 5
	for {
		vals, e := s.Client.ZRangeByScore(s.ctx, "Machine", &redis.ZRangeBy{
			Min: strconv.Itoa(offset),
			Max: strconv.Itoa(offset + interval),
		}).Result()

		if e != nil {
			err = e
			return
		}
		for _, v := range vals {
			vStrArr := strings.Split(v, ",")
			if len(vStrArr) == 3 {
				if c, _ := strconv.ParseFloat(vStrArr[0], 64); err == nil {
					cpu = append(cpu, c)
				}
				if m, _ := strconv.ParseFloat(vStrArr[1], 64); err == nil {
					mem = append(mem, m)
				}
				if d, _ := strconv.ParseFloat(vStrArr[2], 64); err == nil {
					disk = append(disk, d)
				}
			}
		}

		if len(vals) < interval {
			return
		} else {
			offset += interval
		}
	}
}

func (s *RedisStatistics) GetPackerStatistics(tickerCount int) (error, *[][]uint64) {
	offset := 0
	interval := 10000
	result := make([][]uint64, 16)
	packerKey := ""
	keyLen := 0

	if config.ConfigObj.ClientConfig.Enable {
		packerKey = "ClientPacker"
		keyLen += 8
	}

	if config.ConfigObj.ServerConfig.Enable {
		packerKey = "ServerPacker"
		keyLen += 8
	}

	for {
		vals, e := s.Client.ZRangeByScore(s.ctx, packerKey, &redis.ZRangeBy{
			Min: strconv.Itoa(offset),
			Max: strconv.Itoa(offset + interval),
		}).Result()

		if e != nil {
			return e, nil
		}

		for _, v := range vals {
			vStrArr := strings.Split(v, ",")
			if len(vStrArr) == keyLen+1 {
				for i, vStr := range vStrArr[0:keyLen] {
					vInt, _ := strconv.Atoi(vStr)
					result[i] = append(result[i], uint64(vInt))
				}
			}
		}

		if len(vals) < interval {
			return nil, &result
		} else {
			offset += interval
		}
	}
}
