package statistics

import (
	"context"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
	"math"
	"mock-cmpp-stress-test/config"
	"os"
	"time"
)

type CollectionService interface {
	Init(logger *zap.Logger, ctx context.Context)
	Start() error
	Stop() error

	SaveMachineStatistics(tickerCount int, cpu, mem, disk float64) error
	SavePackerStatistics(tickerCount int) error
	AddPackerStatistics(name string, success bool)

	GetXAxisStart(tickerCount int) int
	GetXAxisLength(tickerCount int) int
	GetMachineStatistics(tickerCount int) (err error, cpu, mem, disk []float64)
	GetPackerStatistics(tickerCount int) (error, *[][]uint64)
}

type Collection struct {
	cfg         *config.RedisConfig
	ctx         context.Context
	cancel      context.CancelFunc
	Logger      *zap.Logger
	Service     CollectionService
	TickerCount int
}

var CollectService *Collection

func (s *Collection) Init(log *zap.Logger) {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.Logger = log
	s.cfg = config.ConfigObj.Redis
	if s.cfg.Enable {
		s.Service = CollectionService(new(RedisStatistics))
	} else {
		s.Service = CollectionService(new(Statistics))
	}
	s.Service.Init(s.Logger, s.ctx)
	CollectService = s
}

func (s *Collection) Start() error {
	if err := s.Service.Start(); err != nil {
		s.Logger.Error("Collect Service Start Error.", zap.Error(err))
	}
	s.Logger.Info("Collect Service Start Success.")
	go s.CollectionStatistics()
	return nil
}

func (s *Collection) Stop() error {
	s.Graph()
	if err := s.Service.Stop(); err != nil {
		s.Logger.Error("Collect Service Stop Error.", zap.Error(err))
	}
	s.cancel()
	s.Logger.Info("Collect Service Stop Success.")
	return nil
}

func (s *Collection) CollectionStatistics() {
	t := time.NewTicker(1 * time.Second)
	s.TickerCount = 0

	for {
		select {
		case <-t.C:
			//s.Logger.Info("[Collect][CollectionStatistics] Start", zap.Int("TickerCount", tickerCount))
			s.SaveMachineStatistics(s.TickerCount)
			s.SavePackerStatistics(s.TickerCount)
			s.TickerCount += 1
		case <-s.ctx.Done():
			t.Stop()
			return
		}
	}
}

func (s *Collection) SaveMachineStatistics(tickerCount int) {
	cpuPercents, _ := cpu.Percent(time.Second, false)
	cpuPercent := cpuPercents[0]

	memInfo, _ := mem.VirtualMemory()
	memPercent := memInfo.UsedPercent

	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	diskPercent := diskInfo.UsedPercent
	err := s.Service.SaveMachineStatistics(tickerCount, cpuPercent, memPercent, diskPercent)
	if err != nil {
		s.Logger.Error("[Collection][GetMachineStatistics] Error",
			zap.Int("TickerCount", tickerCount),
			zap.Float64("CpuPercent", cpuPercent),
			zap.Float64("MemPercent", memPercent),
			zap.Float64("DiskPercent", diskPercent),
			zap.Error(err))
	}
}

func (s *Collection) SavePackerStatistics(tickerCount int) {
	if err := s.Service.SavePackerStatistics(tickerCount); err != nil {
		s.Logger.Error("[Collect][SavePackerStatistics] Error",
			zap.Error(err))
	}
}

func (s *Collection) Graph() {
	s.GraphMachine()
	s.GraphPackage()
}

func (s *Collection) GraphMachine() {
	err, cpuData, memData, diskData := s.Service.GetMachineStatistics(s.TickerCount)
	if err != nil {
		s.Logger.Error("[Collect][GraphMachine] Error", zap.Error(err))
		return
	}

	xStart := s.Service.GetXAxisStart(s.TickerCount)
	xLen := s.Service.GetXAxisLength(s.TickerCount)
	xAxis := GetXAxis(xStart, s.TickerCount, xLen)
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Mock CMPP Stress Test",
			Subtitle: "Machine Usage",
		}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true}),
		charts.WithYAxisOpts(opts.YAxis{
			Show:  true,
			Scale: true,
			Min:   0,
			Max:   150,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:         true,
			SelectedMode: "multiple",
			Bottom:       "0",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
			//Trigger:   "item",
			TriggerOn: "mousemove",
		}),
	)

	var markPoints = []charts.SeriesOpts{
		charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{
			Name: "最大值",
			Type: "max",
		}),
		charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{
			Name: "平均值",
			Type: "average",
		}),
		charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{
			Name: "最小值",
			Type: "min",
		}),
	}
	line.SetXAxis(xAxis).
		AddSeries("CPU", GetLineFloatItem(cpuData), markPoints...).
		AddSeries("Memory", GetLineFloatItem(memData), markPoints...).
		AddSeries("Disk", GetLineFloatItem(diskData), markPoints...).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	w, _ := os.Create("CMPP_Stress_Test_Machine.html")
	renderErr := line.Render(w)
	if renderErr != nil {
		s.Logger.Error("[Collect][GraphMachine] Render Error", zap.Error(renderErr))
		return
	}
}

func (s *Collection) GraphPackage() {
	err, data := s.Service.GetPackerStatistics(s.TickerCount)
	if err != nil {
		s.Logger.Error("[Collect][GraphPackage] Error", zap.Error(err))
		return
	}

	if len(*data) == 0 {
		return
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeWesteros,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Mock CMPP Stress Test",
			Subtitle: "Package",
		}),
		//charts.WithToolboxOpts(opts.Toolbox{Show: true}),
		charts.WithYAxisOpts(opts.YAxis{
			Show:  true,
			Scale: true,
			Max:   int(float64((*data)[0][len((*data)[0])-1]) * 1.1), // 为了美观
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:         true,
			SelectedMode: "multiple",
			Bottom:       "-5",
			Orient:       "horizontal",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:      true,
			Trigger:   "item",
			TriggerOn: "mousemove",
		}),
	)

	xStart := s.Service.GetXAxisStart(s.TickerCount)
	xLen := s.Service.GetXAxisLength(s.TickerCount)
	xAxis := GetXAxis(xStart, s.TickerCount, xLen)

	var markPoints = []charts.SeriesOpts{
		charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{
			Name: "最大值",
			Type: "max",
		}),
		charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{
			Name: "平均值",
			Type: "average",
		}),
		charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{
			Name: "最小值",
			Type: "min",
		}),
	}
	line.SetXAxis(xAxis).
		AddSeries("Submit Total Package", GetLineUintItem((*data)[0]), markPoints...).
		AddSeries("Submit Success Package", GetLineUintItem((*data)[1])).
		AddSeries("Submit Resp Total Package", GetLineUintItem((*data)[2])).
		AddSeries("Submit Resp Success Package", GetLineUintItem((*data)[3])).
		AddSeries("Deliver Total Package", GetLineUintItem((*data)[4])).
		AddSeries("Deliver Success Package", GetLineUintItem((*data)[5])).
		AddSeries("Deliver Resp Total Package", GetLineUintItem((*data)[6])).
		AddSeries("Deliver Resp Success Package", GetLineUintItem((*data)[7])).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	f, _ := os.Create("CMPP_Stress_Test_Package.html")
	renderErr := line.Render(f)
	if renderErr != nil {
		s.Logger.Error("[Collect][GraphPackage] Render Error", zap.Error(renderErr))
		return
	}

}

func GetXAxis(start, end, len int) []int {
	xAxis := make([]int, 0)
	multiple := end / len
	for i := start; i < len; i++ {
		if multiple != 0 {
			xAxis = append(xAxis, i+len*(multiple-1))
		} else {
			xAxis = append(xAxis, i)
		}
	}

	if start != 0 {
		for i := 0; i < start; i++ {
			xAxis = append(xAxis, i+len*multiple)
		}
	}

	return xAxis
}

func GetLineFloatItem(data []float64) []opts.LineData {
	result := make([]opts.LineData, len(data))
	for i, d := range data {
		result[i] = opts.LineData{Value: math.Floor(d*100) / 100}
	}
	return result
}

func GetLineUintItem(data []uint64) []opts.LineData {
	result := make([]opts.LineData, len(data))
	for i, d := range data {
		result[i] = opts.LineData{Value: d}
	}
	return result
}
