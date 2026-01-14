package cmd

import (
	"Warden/pb"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedCitadelServer
}

func (s *server) Watch(stream pb.Citadel_WatchServer) error {
	for {
		report, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Command{
				Action: "KEEP_ALIVE",
			})
		}
		if err != nil {
			return err
		}

		formattedTime := report.GetTimestamp().AsTime().Format("2006-01-02 15:04:05")
		uptime := time.Duration(report.GetUptime()) * time.Second
		netSentKB := float64(report.GetNetSent()) / 1024
		netRecvKB := float64(report.GetNetRecv()) / 1024

		fmt.Printf("[%v] %s: CPU %.2f%%, MEM %.2f%%, DISK %.2f%%, NET SENT %.2f KB, NET RECV %.2f KB, UPTIME %v\n",
			formattedTime,
			report.GetWardenId(),
			report.GetCpuUsage(),
			report.GetMemUsage(),
			report.GetDiskUsage(),
			netSentKB,
			netRecvKB,
			uptime)
	}
}

var citadelCmd = &cobra.Command{
	Use:   "Citadel",
	Short: "Citadel 모니터링 서버를 실행합니다.",
	Run: func(cmd *cobra.Command, args []string) {
		port := os.Getenv("CITADEL_PORT")
		if port == "" {
			log.Fatal("Not found CITADEL_PORT in environment config")
		}

		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		// gRPC 서버 인스턴스 생성
		s := grpc.NewServer()

		// 서비스 등록
		pb.RegisterCitadelServer(s, &server{})

		log.Printf("Citadel server listening on: %s...", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(citadelCmd)
}
