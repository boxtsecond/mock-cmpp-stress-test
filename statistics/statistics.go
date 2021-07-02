package statistics

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"mock-cmpp-stress-test/config"
	"sync/atomic"
)

type Statistics struct {
	Logger            *zap.Logger
	Machine           *MachineStatistics
	ClientSubmit      *PackerStatistics
	ClientSubmitResp  *PackerStatistics
	ClientDeliver     *PackerStatistics
	ClientDeliverResp *PackerStatistics

	ServerSubmit      *PackerStatistics
	ServerSubmitResp  *PackerStatistics
	ServerDeliver     *PackerStatistics
	ServerDeliverResp *PackerStatistics
}

type MachineStatistics struct {
	Item
	Statistics [][]float64
}

type PackerStatistics struct {
	Item
	Statistics [][]uint64
}

type Item struct {
	MaxStatisticsCount int
	Head               uint
	Total              uint64
	Success            uint64
}

const DefaultMaxStatisticsCount = 30 * 60

func (s *Statistics) Init(log *zap.Logger, ctx context.Context) {
	//s.ctx = ctx
	s.Logger = log
	s.Machine = &MachineStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]float64, DefaultMaxStatisticsCount),
	}
	s.ClientSubmit = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}
	s.ClientSubmitResp = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}
	s.ClientDeliver = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}
	s.ClientDeliverResp = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}

	s.ServerSubmit = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}
	s.ServerSubmitResp = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}
	s.ServerDeliver = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}
	s.ServerDeliverResp = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, DefaultMaxStatisticsCount),
	}
}

func (s *Statistics) Start() error {
	return nil
}

func (s *Statistics) Stop() error {
	return nil
}

func (s *Statistics) NewDefaultItem() Item {
	return Item{
		MaxStatisticsCount: DefaultMaxStatisticsCount,
		Head:               0,
		Total:              0,
		Success:            0,
	}
}

func (s *Statistics) SaveMachineStatistics(tickerCount int, cpu, mem, disk float64) error {
	index := tickerCount
	if tickerCount >= s.Machine.MaxStatisticsCount {
		index = tickerCount % (s.Machine.MaxStatisticsCount)
		s.Machine.Head = uint((index + 1) % s.Machine.MaxStatisticsCount)
	}

	s.Machine.Statistics[index] = []float64{cpu, mem, disk}
	return nil
}

func (s *Statistics) SavePackerStatistics(tickerCount int) error {
	s.ClientSubmit.SavePackerStatistics(tickerCount)
	s.ClientSubmitResp.SavePackerStatistics(tickerCount)
	s.ClientDeliver.SavePackerStatistics(tickerCount)
	s.ClientDeliverResp.SavePackerStatistics(tickerCount)

	s.ServerSubmit.SavePackerStatistics(tickerCount)
	s.ServerSubmitResp.SavePackerStatistics(tickerCount)
	s.ServerDeliver.SavePackerStatistics(tickerCount)
	s.ServerDeliverResp.SavePackerStatistics(tickerCount)
	return nil
}

func (ps *PackerStatistics) SavePackerStatistics(tickerCount int) {
	index := tickerCount
	if tickerCount >= ps.MaxStatisticsCount {
		index = tickerCount % (ps.MaxStatisticsCount)
		ps.Head = uint((index + 1) % ps.MaxStatisticsCount)
	}
	ps.Statistics[index] = []uint64{ps.Total, ps.Success}
}

func (s *Statistics) AddPackerStatistics(source, name string, success bool) {
	key := fmt.Sprintf("%s_%s", source, name)
	switch key {
	case "Client_Submit":
		s.ClientSubmit.AddPackerStatistics(success)
	case "Client_SubmitResp":
		s.ClientSubmitResp.AddPackerStatistics(success)
	case "Client_Deliver":
		s.ClientDeliver.AddPackerStatistics(success)
	case "Client_DeliverResp":
		s.ClientDeliverResp.AddPackerStatistics(success)

	case "Server_Submit":
		s.ServerSubmit.AddPackerStatistics(success)
	case "Server_SubmitResp":
		s.ServerSubmitResp.AddPackerStatistics(success)
	case "Server_Deliver":
		s.ServerDeliver.AddPackerStatistics(success)
	case "Server_DeliverResp":
		s.ServerDeliverResp.AddPackerStatistics(success)
	}
}

func (ps *PackerStatistics) AddPackerStatistics(success bool) {
	if success {
		atomic.AddUint64(&ps.Success, 1)
	}
	atomic.AddUint64(&ps.Total, 1)
}

func (s *Statistics) GetXAxisStart(tickerCount int) int {
	if config.ConfigObj.ClientConfig.Enable {
		if tickerCount <= s.ClientSubmit.MaxStatisticsCount {
			return 0
		}
		return int(s.ClientSubmit.Head)
	}

	if config.ConfigObj.ServerConfig.Enable {
		if tickerCount <= s.ServerSubmit.MaxStatisticsCount {
			return 0
		}
		return int(s.ServerSubmit.Head)
	}
	return 0
}

func (s *Statistics) GetXAxisLength(tickerCount int) int {
	if tickerCount <= s.Machine.MaxStatisticsCount {
		return tickerCount
	}
	return s.Machine.MaxStatisticsCount
}

func (s *Statistics) GetMachineStatistics(tickerCount int) (err error, cpu, mem, disk []float64) {
	start := 0
	total := tickerCount
	if s.Machine.Head != 0 {
		start = int(s.Machine.Head)
		total = s.Machine.MaxStatisticsCount
	}

	for total >= 0 {
		index := (start + s.Machine.MaxStatisticsCount) % s.Machine.MaxStatisticsCount
		stat := s.Machine.Statistics[index]
		if len(stat) == 3 {
			cpu = append(cpu, stat[0])
			mem = append(mem, stat[1])
			disk = append(disk, stat[2])
		}
		total -= 1
		start += 1
	}
	return
}

func (s *Statistics) GetPackerStatistics(tickerCount int) (error, *[][]uint64) {
	start := 0
	total := tickerCount
	result := make([][]uint64, 16)

	if config.ConfigObj.ServerConfig.Enable {
		if s.ServerSubmit.Head != 0 {
			start = int(s.ServerSubmit.Head)
			total = s.ServerSubmit.MaxStatisticsCount
		}
	}

	if config.ConfigObj.ClientConfig.Enable {
		if s.ClientSubmit.Head != 0 {
			start = int(s.ClientSubmit.Head)
			total = s.ClientSubmit.MaxStatisticsCount
		}
	}

	for total > 0 {
		index := (start + s.ClientSubmit.MaxStatisticsCount) % s.ClientSubmit.MaxStatisticsCount
		result[0] = append(result[0], s.ClientSubmit.Statistics[index][0])
		result[1] = append(result[1], s.ClientSubmit.Statistics[index][1])

		result[2] = append(result[2], s.ClientSubmitResp.Statistics[index][0])
		result[3] = append(result[3], s.ClientSubmitResp.Statistics[index][1])

		result[4] = append(result[4], s.ClientDeliver.Statistics[index][0])
		result[5] = append(result[5], s.ClientDeliver.Statistics[index][1])

		result[6] = append(result[6], s.ClientDeliverResp.Statistics[index][0])
		result[7] = append(result[7], s.ClientDeliverResp.Statistics[index][1])

		result[8] = append(result[8], s.ServerSubmit.Statistics[index][0])
		result[9] = append(result[9], s.ServerSubmit.Statistics[index][1])

		result[10] = append(result[10], s.ServerSubmitResp.Statistics[index][0])
		result[11] = append(result[11], s.ServerSubmitResp.Statistics[index][1])

		result[12] = append(result[12], s.ServerDeliver.Statistics[index][0])
		result[13] = append(result[13], s.ServerDeliver.Statistics[index][1])

		result[14] = append(result[14], s.ServerDeliverResp.Statistics[index][0])
		result[15] = append(result[15], s.ServerDeliverResp.Statistics[index][1])
		total -= 1
		start += 1
	}

	return nil, &result
}
