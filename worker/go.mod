module github.com/jkfastdevth/Siam-Synapse/worker

go 1.25.7

replace github.com/jkfastdevth/Siam-Synapse/proto => ../proto

require (
	github.com/jkfastdevth/Siam-Synapse/proto v0.0.0-00010101000000-000000000000
	github.com/shirou/gopsutil v3.21.11+incompatible
	google.golang.org/grpc v1.79.1
)

require (
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
