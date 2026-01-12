package cmd

import (
	"Warden/pb"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type client struct {
	warden pb.CitadelClient
	id     string
}

var wg sync.WaitGroup

var wardenCmd = &cobra.Command{
	Use:   "Warden",
	Short: "Warden이 Citadel에게 모니터링 정보를 주기적으로 보고합니다.",
	Run: func(cmd *cobra.Command, args []string) {
		server, _ := cmd.Flags().GetString("server")
		port, _ := cmd.Flags().GetString("port")
		id, _ := cmd.Flags().GetString("id")

		// 종료 시그널을 감지할 컨텍스트 생성
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// 보안 연결(TLS/SSL) 미설정
		conn, err := grpc.NewClient(server+":"+port,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("서버에 연결할 수 없습니다.: %v", err)
		}
		defer conn.Close()
		defer wg.Wait()

		// Client 생성
		c := &client{
			warden: pb.NewCitadelClient(conn),
			id:     id,
		}
		stream, err := c.warden.Watch(ctx)
		if err != nil {
			log.Fatalf("스트림을 열 수 없습니다. %v", err)
		}

		// 정보 수집
		cpuUsage := make(chan float32)
		memUsage := make(chan float32)
		uptime := make(chan int64)

		for {
			go func() { cpuUsage <- getCpuUsage() }()
			go func() { memUsage <- getMemUsage() }()
			go func() { uptime <- getUptime() }()

			info := &pb.Report{
				WardenId:  c.id,
				Timestamp: timestamppb.Now(),
				CpuUsage:  <-cpuUsage,
				MemUsage:  <-memUsage,
				Uptime:    <-uptime,
			}
			if err := stream.Send(info); err != nil {
				select {
				case <-ctx.Done():
					log.Printf("%s을 종료합니다. 리소스를 정리 중...", id)
					stream.CloseSend()
					// Run 함수 종료 ->> defer conn.Close() 및 wg.Wait() 실행
					return
				default:
					log.Printf("전송 실패: %v", err)
					return
				}
			}
			log.Printf("Sent - %v\n", info)
		}
	},
}

func init() {
	wardenCmd.Flags().StringP("server", "s", "localhost", "Citadel server address")
	wardenCmd.Flags().StringP("port", "p", "50051", "Citadel server port")
	wardenCmd.Flags().StringP("id", "i", "default-warden", "Unique ID for this Warden")
	rootCmd.AddCommand(wardenCmd)
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
