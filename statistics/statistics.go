package statistics

import (
	"context"
	"go.uber.org/zap"
	"sync/atomic"
)

type Statistics struct {
	Logger      *zap.Logger
	Machine     *MachineStatistics
	Submit      *PackerStatistics
	SubmitResp  *PackerStatistics
	Deliver     *PackerStatistics
	DeliverResp *PackerStatistics
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

func (s *Statistics) Init(log *zap.Logger, ctx context.Context) {
	//s.ctx = ctx
	s.Logger = log
	s.Machine = &MachineStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]float64, 30*60),
	}
	s.Submit = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, 30*60),
	}
	s.SubmitResp = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, 30*60),
	}
	s.Deliver = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, 30*60),
	}
	s.DeliverResp = &PackerStatistics{
		Item:       s.NewDefaultItem(),
		Statistics: make([][]uint64, 30*60),
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
		MaxStatisticsCount: 30 * 60,
		Head:               0,
		Total:              0,
		Success:            0,
	}
}

func (s *Statistics) SaveMachineStatistics(tickerCount int, cpu, mem, disk float64) error {
	index := tickerCount
	if tickerCount > s.Machine.MaxStatisticsCount {
		index = tickerCount % (s.Machine.MaxStatisticsCount + 1)
		s.Machine.Head = uint(index) + 1
	}

	s.Machine.Statistics[index] = []float64{cpu, mem, disk}
	return nil
}

func (s *Statistics) SavePackerStatistics(tickerCount int) error {
	s.Submit.SavePackerStatistics(tickerCount)
	s.SubmitResp.SavePackerStatistics(tickerCount)
	s.Deliver.SavePackerStatistics(tickerCount)
	s.DeliverResp.SavePackerStatistics(tickerCount)
	return nil
}

func (ps *PackerStatistics) SavePackerStatistics(tickerCount int) {
	index := tickerCount
	if tickerCount > ps.MaxStatisticsCount {
		index = tickerCount % (ps.MaxStatisticsCount + 1)
		ps.Head = uint(index) + 1
	}
	ps.Statistics[index] = []uint64{ps.Total, ps.Success}
}

func (s *Statistics) AddPackerStatistics(name string, success bool) {
	switch name {
	case "Submit":
		s.Submit.AddPackerStatistics(success)
	case "SubmitResp":
		s.SubmitResp.AddPackerStatistics(success)
	case "Deliver":
		s.Deliver.AddPackerStatistics(success)
	case "DeliverResp":
		s.DeliverResp.AddPackerStatistics(success)
	}
}

func (ps *PackerStatistics) AddPackerStatistics(success bool) {
	if success {
		atomic.AddUint64(&ps.Success, 1)
	}
	atomic.AddUint64(&ps.Total, 1)
}

func (s *Statistics) GetXAxisStart(tickerCount int) int {
	if tickerCount < s.Submit.MaxStatisticsCount {
		return 0
	}
	return int(s.Submit.Head)
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
	result := make([][]uint64, 8)

	if s.Submit.Head != 0 {
		start = int(s.Submit.Head)
		total = s.Submit.MaxStatisticsCount
	}

	for total > 0 {
		index := (start + s.Submit.MaxStatisticsCount) % s.Submit.MaxStatisticsCount
		result[0] = append(result[0], s.Submit.Statistics[index][0])
		result[1] = append(result[1], s.Submit.Statistics[index][1])

		result[2] = append(result[2], s.SubmitResp.Statistics[index][0])
		result[3] = append(result[3], s.SubmitResp.Statistics[index][1])

		result[4] = append(result[4], s.Deliver.Statistics[index][0])
		result[5] = append(result[5], s.Deliver.Statistics[index][1])

		result[6] = append(result[6], s.DeliverResp.Statistics[index][0])
		result[7] = append(result[7], s.DeliverResp.Statistics[index][1])
		total -= 1
		start += 1
	}

	return nil, &result
}
