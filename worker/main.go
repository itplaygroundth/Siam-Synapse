// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"
// 	"net"
// 	 //"io"
// 	"os"
// 	"os/exec"
// 	"github.com/shirou/gopsutil/v3/cpu"
// 	"github.com/shirou/gopsutil/v3/mem"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// 	"github.com/jkfastdevth/Siam-Synapse/proto"
// )

// type workerServer struct {
// 	proto.UnimplementedMasterControlServer
// 	nodeID string
// }

// func restartContainer(containerName string) error {
// 	log.Printf("🐳 Executing: docker restart %s", containerName)
// 	cmd := exec.Command("docker", "restart", containerName)
// 	return cmd.Run()
// }

// // Implement ManageContainer ใน Worker
// func (s *workerServer) ManageContainer(ctx context.Context, req *proto.ContainerAction) (*proto.Ack, error) {
// 	log.Printf("🛠️ Received Container Action: %s on %s", req.Action, req.ContainerName)

// 	if req.Action == "restart" {
// 		err := restartContainer(req.ContainerName)
// 		if err != nil {
// 			log.Printf("❌ Failed to restart container: %v", err)
// 			return &proto.Ack{Success: false, Message: err.Error()}, nil
// 		}
// 		log.Printf("✅ Container %s restarted successfully", req.ContainerName)
// 		return &proto.Ack{Success: true, Message: "Container restarted"}, nil
// 	}

// 	return &proto.Ack{Success: false, Message: "Unknown action"}, nil
// }


// // 1. ฟังก์ชันหาชื่อ Container ตัวเอง
// func getContainerName() string {
//     // ใช้ hostname ของ container มักจะเป็น ID หรือชื่อที่ Docker ตั้งให้
// 	hostname, err := os.Hostname()
// 	if err != nil {
// 		return "unknown"
// 	}
// 	return hostname
// }


// func main() {

	
// 	conn, err := grpc.Dial("master:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	client := proto.NewMasterControlClient(conn)
	
	
// 	if err != nil {
// 		log.Fatalf("❌ Could not connect to Master: %v", err)
// 	}
// 	defer conn.Close()

// 	stream, err := client.GetCommand(context.Background(), &proto.NodeStatus{NodeId: "worker-ubuntu-01"})
//     if err != nil {
//         log.Fatalf("❌ Failed to get command stream: %v", err)
//     }
//     log.Println("📡 Stream opened")

// 	go func() {
// 			lis, err := net.Listen("tcp", ":50052")
// 			if err != nil {
// 				log.Fatalf("failed to listen: %v", err)
// 			}
// 			s := grpc.NewServer()
// 			proto.RegisterMasterControlServer(s, &workerServer{
// 				nodeID: os.Getenv("NODE_ID"), // ใช้ nodeID จาก env
// 			})
// 			log.Println("🛠️ Worker gRPC Server listening on :50052")
// 			if err := s.Serve(lis); err != nil {
// 				log.Fatalf("failed to serve: %v", err)
// 			}
// 		}()

// 	// 1. เชื่อมต่อกับ Master (Port 50051)
// 	// ใน Docker Compose ให้ใช้ชื่อ service "master" แทน IP


// 	grpcClient := proto.NewMasterControlClient(conn)

// 	fmt.Println("🚀 Worker Node Started: Sending heartbeats to Master...")


// // 2. จำลอง Node ID
// 	nodeID := os.Getenv("NODE_ID")
// 	if nodeID == "" {
// 		nodeID = "worker-ubuntu-01"
// 	}
// 	// 3. --- เพิ่มส่วนนี้ ---
// 	// คอยฟังคำสั่ง (Stream)
// 	stream, err := grpcClient.GetCommand(context.Background(), &proto.NodeStatus{NodeId: nodeID})
// 	if err != nil {
// 		log.Fatalf("could not get command: %v", err)
// 	}

	

// 	go func() {
// 		for {
// 			cmd, err := stream.Recv()
// 		if err != nil {
//             log.Printf("❌ Error receiving command: %v", err)
//             break
//         }
// 		fmt.Printf("📥 Received Command: %s\n", cmd.Command)
// 			// if err != nil {
// 			// 	log.Fatalf("Error receiving command: %v", err)
// 			// }

// 			// fmt.Printf("📥 Received Command: %s\n", cmd.Command)
// 			if cmd.Command == "restart" {
//             // ใช้ชื่อจริงที่คุณตั้งใน docker-compose: sworker-ubuntu-01
//             err := restartContainer("sworker-ubuntu-01") 
//            if cmd.Command == "restart" {
//                 restartContainer("sworker-ubuntu-01") 
//             }
//         }
// 			// ตรงนี้คือ Logic ที่จะจัดการคำสั่ง เช่นสั่ง restart หรือ stop
// 		}
// 	}()
// 	// -----------------

 

// 	// 2. Loop ส่ง Heartbeat ทุกๆ 5 วินาที
// 	for {
// 		// ดึงข้อมูลระบบจริง
// 		c, _ := cpu.Percent(0, false)
// 		m, _ := mem.VirtualMemory()

// 		status := &proto.NodeStatus{
// 			NodeId:   "worker-ubuntu-01",
// 			CpuUsage: float32(c[0]),
// 			RamUsage: float32(m.UsedPercent),
// 			Status:   "Online",
// 		}

// 		// ยิง gRPC ไปที่ Master
// 		res, err := grpcClient.ReportStatus(context.Background(), status)
// 		if err != nil {
// 			fmt.Printf("⚠️ Error sending status: %v\n", err)
// 		} else {
// 			fmt.Printf("✅ Master Response: %s (CPU: %.1f%%)\n", res.Message, status.CpuUsage)
// 		}

// 		time.Sleep(5 * time.Second)
// 	}
// }

package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/shirou/gopsutil/cpu" // ต้องติดตั้งเพิ่ม: go get github.com/shirou/gopsutil/cpu
    "github.com/shirou/gopsutil/mem" // ต้องติดตั้งเพิ่ม: go get github.com/shirou/gopsutil/mem
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "github.com/jkfastdevth/Siam-Synapse/proto" // เปลี่ยนเป็น path โปรเจคของคุณ
)

// จำลองฟังก์ชัน restartContainer
func restartContainer(containerName string) error {
    log.Printf("🐳 Executing: docker restart %s", containerName)
    // ตรงนี้คือ logic สั่ง docker จริงๆ
    return nil
}

func main() {
    // 1. จำลอง Node ID จาก ENV
    nodeID := os.Getenv("NODE_ID")
    if nodeID == "" {
        nodeID = "worker-ubuntu-01"
    }

    // 2. เชื่อมต่อกับ Master (Port 50051)
    conn, err := grpc.Dial("master:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("❌ Could not connect to Master: %v", err)
    }
    defer conn.Close()
    
    grpcClient := proto.NewMasterControlClient(conn)
    
    // 3. เปิด Stream รับคำสั่งจาก Master
    stream, err := grpcClient.GetCommand(context.Background(), &proto.NodeStatus{NodeId: nodeID})
    if err != nil {
        log.Fatalf("❌ Failed to get command stream: %v", err)
    }
    log.Println("📡 Stream opened")

    // 4. Goroutine รับคำสั่ง
    go func() {
        for {
            cmd, err := stream.Recv()
            if err != nil {
                log.Printf("❌ Error receiving command: %v", err)
                return // ออกจาก goroutine
            }
            fmt.Printf("📥 Received Command: %s\n", cmd.Command)
            
            if cmd.Command == "restart" {
                // ใช้ชื่อจริงที่คุณตั้งใน docker-compose: sworker-ubuntu-01
                err := restartContainer("sworker-ubuntu-01") 
                if err != nil {
                    log.Printf("❌ Failed to restart: %v", err)
                }
            }
        }
    }()

    // 5. Loop ส่ง Heartbeat ทุกๆ 5 วินาที
    fmt.Println("🚀 Worker Node Started: Sending heartbeats to Master...")
    for {
        c, _ := cpu.Percent(0, false)
        m, _ := mem.VirtualMemory()

        status := &proto.NodeStatus{
            NodeId:   nodeID, // ใช้ nodeID ที่ได้จาก ENV
            CpuUsage: float32(c[0]),
            RamUsage: float32(m.UsedPercent),
            Status:   "Online",
        }

        res, err := grpcClient.ReportStatus(context.Background(), status)
        if err != nil {
            fmt.Printf("⚠️ Error sending status: %v\n", err)
        } else {
            fmt.Printf("✅ Master Response: %s (CPU: %.1f%%)\n", res.Message, status.CpuUsage)
        }

        time.Sleep(5 * time.Second)
    }
}