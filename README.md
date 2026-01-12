# Warden 🛡️

Warden은 Go 언어로 작성된 고성능 모니터링 에이전트입니다. 여러 호스트에서 시스템 메트릭과 애플리케이션 성능 데이터를 수집하여 중앙 서버(Citadel)로 gRPC를 통해 실시간 스트리밍하는 것을 목표로 합니다.

## Core Objectives 🎯

- **실시간 성능 데이터 수집**: CPU, Memory, Disk, Network 메트릭 수집.
- **시스템 부하 최소화**: 에이전트의 CPU 점유율을 1% 미만으로 유지하는 것을 목표로 설계.
- **고성능 통신**: gRPC 기반의 확장 가능한 클라이언트-사이드 스트리밍 아키텍처.
- **APM 기능**: End-to-End 트랜잭션 추적 및 지연 시간/에러 측정.

## Technical Stack 🛠️

- **Language**: Go (Goroutine/Channel 동시성 활용)
- **Communication**: gRPC / Protocol Buffers (Binary streaming)
- **CLI Framework**: Cobra
- **System Metrics**: `gopsutil` / Linux `/proc` filesystem
