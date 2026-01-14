package collector

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Metrics struct {
	CpuUsage  float32
	MemUsage  float32
	Uptime    int64
	NetSent   uint64
	NetRecv   uint64
	DiskUsage float32
}

var wg sync.WaitGroup

var isFirstMetric bool = true
var lastNetSent uint64
var lastNetRecv uint64

func GetHostResources() Metrics {
	var m Metrics

	cpuChan := make(chan float32, 1)
	memChan := make(chan float32, 1)
	uptimeChan := make(chan int64, 1)
	netSentChan := make(chan uint64, 1)
	netRecvChan := make(chan uint64, 1)
	diskChan := make(chan float32, 1)

	defer wg.Wait()
	defer close(cpuChan)
	defer close(memChan)
	defer close(uptimeChan)
	defer close(netSentChan)
	defer close(netRecvChan)
	defer close(diskChan)

	go func() { cpuChan <- getCpuUsage() }()
	go func() { memChan <- getMemUsage() }()
	go func() { uptimeChan <- getUptime() }()
	go func() {
		sent, recv := getNetInfo()
		netSentChan <- sent
		netRecvChan <- recv
	}()
	go func() { diskChan <- getDiskUsage() }()

	m.CpuUsage = <-cpuChan
	m.MemUsage = <-memChan
	m.Uptime = <-uptimeChan
	m.NetSent = <-netSentChan
	m.NetRecv = <-netRecvChan
	m.DiskUsage = <-diskChan

	return m
}

func getCpuUsage() float32 {
	wg.Add(1)
	defer wg.Done()
	cpuInfo, _ := cpu.Percent(time.Second, false)
	return float32(cpuInfo[0])
}

func getMemUsage() float32 {
	wg.Add(1)
	defer wg.Done()
	mem, _ := mem.VirtualMemory()
	return float32(mem.UsedPercent)
}

func getUptime() int64 {
	wg.Add(1)
	defer wg.Done()
	uptime, _ := host.Uptime()
	return int64(uptime)
}

func getNetInfo() (uint64, uint64) {
	wg.Add(1)
	defer wg.Done()
	io, _ := net.IOCounters(false)
	if len(io) == 0 {
		return 0, 0
	}

	if isFirstMetric {
		isFirstMetric = false
		lastNetSent = io[0].BytesSent
		lastNetRecv = io[0].BytesRecv
		return 0, 0
	}

	currSent := io[0].BytesSent
	currRecv := io[0].BytesRecv

	diffNetSent := currSent - lastNetSent
	diffNetRecv := currRecv - lastNetRecv

	lastNetSent = currSent
	lastNetRecv = currRecv

	return diffNetSent, diffNetRecv
}

func getDiskUsage() float32 {
	wg.Add(1)
	defer wg.Done()
	usage, _ := disk.Usage("/")
	return float32(usage.UsedPercent)
}
