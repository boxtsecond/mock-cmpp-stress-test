package stress_test_service

import (
	"context"
	"errors"
	cmpp "github.com/bigwhite/gocmpp"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/cmpp/client"
	"mock-cmpp-stress-test/config"
	"mock-cmpp-stress-test/statistics"
	"sync"
	"sync/atomic"
	"time"
)

const defaultConcurrency = 1000

type StressTest struct {
	cfg    *config.StressTestConfig
	Logger *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

func (st *StressTest) Init(log *zap.Logger) {
	st.cfg = config.ConfigObj.StressTest
	st.Logger = log
	st.ctx, st.cancel = context.WithCancel(context.Background())
}

func (st *StressTest) Start() error {
	if !st.cfg.Enable {
		return nil
	}

	if len(client.Clients) == 0 {
		err := errors.New("cmpp clients have no available")
		st.Logger.Error("Stress Test Start Error", zap.Error(err))
		return err
	}

	for _, worker := range *st.cfg.Workers {
		if worker.DurationTime == 0 && worker.TotalNum == 0 {
			err := errors.New("DurationTime and TotalNum can't be 0 at once")
			st.Logger.Error("Stress Test Worker Config Error", zap.Error(err))
			return err
		}

		if worker.TotalNum != 0 {
			go st.StartWorkerByTotalNum(&worker)
			continue
		}

		if worker.DurationTime != 0 {
			go st.StartWorkerByDurationTime(&worker)
			continue
		}
	}

	return nil
}

func (st *StressTest) Stop() error {
	st.cancel()
	st.Logger.Info("Stress Test Stop Success")
	return nil
}

func (st *StressTest) StartWorkerByDurationTime(worker *config.StressTestWorker) {
	cmppClient, ok := client.Clients[worker.Name]
	if !ok {
		st.Logger.Error("[StressTest][StartWorkerByDurationTime] Error", zap.Error(errors.New("can't find cmpp client")))
		return
	}

	concurrency := worker.Concurrency
	// 每个 worker 每秒最多发送 1000 个数据包，避免发不完，达不到并发量
	if worker.Concurrency > defaultConcurrency {
		concurrency = defaultConcurrency
	}
	workerNum := worker.Concurrency / concurrency
	if workerNum == 0 {
		workerNum = 1
	}
	count := uint64(0)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for count < worker.DurationTime {
		select {
		case <-ticker.C:
			for i := uint64(0); i < workerNum; i++ {
				go func(id uint64) {
					for sendNum := uint64(0); sendNum < concurrency; sendNum++ {
						for _, msg := range *st.cfg.Messages {
							if cmppClient.Version == cmpp.V20 || cmppClient.Version == cmpp.V21 {
								err, _ := cmppClient.Cmpp2Submit(&msg)
								if err != nil {
									statistics.CollectService.Service.AddPackerStatistics("Submit", false)
								} else {
									statistics.CollectService.Service.AddPackerStatistics("Submit", true)
								}
							} else if cmppClient.Version == cmpp.V30 {
								err, _ := cmppClient.Cmpp3Submit(&msg)
								if err != nil {
									statistics.CollectService.Service.AddPackerStatistics("Submit", false)
								} else {
									statistics.CollectService.Service.AddPackerStatistics("Submit", true)
								}
							}
						}
						if worker.Sleep > 0 {
							time.Sleep(time.Duration(worker.Sleep) * time.Millisecond)
						}
					}
				}(i)
			}
			atomic.AddUint64(&count, 1)
		case <-st.ctx.Done(): // if cancel() execute
			return
		}
	}
}

func (st *StressTest) StartWorkerByTotalNum(worker *config.StressTestWorker) {
	cmppClient, ok := client.Clients[worker.Name]
	if !ok {
		st.Logger.Error("[StressTest][StartWorkerByTotalNum] Error", zap.Error(errors.New("can't find cmpp client")))
		return
	}

	concurrency := worker.Concurrency
	// 每个 worker 每秒最多发送 1000 个数据包，避免发不完，达不到并发量
	if worker.Concurrency > defaultConcurrency {
		concurrency = defaultConcurrency
	}
	workerNum := worker.Concurrency / concurrency
	if workerNum == 0 {
		workerNum = 1
	}
	total := uint64(0)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var mutex sync.Mutex

	for total < worker.TotalNum {
		select {
		case <-ticker.C:
			st.Logger.Info("Stress Test Ticker Total", zap.Uint64("Total", total))
			for i := uint64(0); i < workerNum; i++ {
				if total >= worker.TotalNum {
					return
				}

				go func(id uint64) {
					for sendNum := uint64(0); sendNum < concurrency; sendNum++ {
						for _, msg := range *st.cfg.Messages {
							mutex.Lock()
							atomic.AddUint64(&total, 1)
							if total >= worker.TotalNum+1 {
								return
							}
							mutex.Unlock()
							st.Logger.Info("Stress Test Worker Start", zap.Uint64("WorkerNum", id), zap.Uint64("Total", total))
							if cmppClient.Version == cmpp.V20 || cmppClient.Version == cmpp.V21 {
								err, _ := cmppClient.Cmpp2Submit(&msg)
								if err != nil {
									statistics.CollectService.Service.AddPackerStatistics("Submit", false)
								} else {
									statistics.CollectService.Service.AddPackerStatistics("Submit", true)
								}
							} else if cmppClient.Version == cmpp.V30 {
								err, _ := cmppClient.Cmpp3Submit(&msg)
								if err != nil {
									statistics.CollectService.Service.AddPackerStatistics("Submit", false)
								} else {
									statistics.CollectService.Service.AddPackerStatistics("Submit", true)
								}
							}
							if worker.Sleep > 0 {
								time.Sleep(time.Duration(worker.Sleep) * time.Millisecond)
							}
						}
					}
				}(i)
			}
		case <-st.ctx.Done():
			return
		}
	}
}
