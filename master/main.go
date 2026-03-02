package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time" // ลบออกเพราะไม่ได้ใช้

	"github.com/gofiber/fiber/v2"
	"github.com/jkfastdevth/Siam-Synapse/proto"
	helper "github.com/jkfastdevth/Siam-Synapse/master/helper"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Postgres driver
	//"github.com/moby/moby/api/types"
	 "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"google.golang.org/grpc"
	"github.com/gofiber/contrib/websocket" // 🛠️ เพิ่ม import
	//"github.com/moby/moby/api/types/network"
)

var db *sql.DB

type MasterApp struct {
	mu            sync.RWMutex
	workerClients map[string]proto.MasterControlClient // เก็บ Client ตาม NodeID
}

var appState = &MasterApp{
	workerClients: make(map[string]proto.MasterControlClient),
}

var (
    workerCpuMap = make(map[string]float64) // เก็บ CPU ล่าสุดแยกตาม NodeID
    statsMu      sync.RWMutex               // ใช้ RWMutex เพื่อประสิทธิภาพที่ดีขึ้น
)

type NodeInfo struct {
	ContainerName string
}

	type LogEntry struct {
			NodeID    string  `json:"node_id"`
			CPUUsage  float64 `json:"cpu_usage"`
			RAMUsage  float64 `json:"ram_usage"`
			Status    string  `json:"status"`
			CreatedAt string  `json:"created_at"`
		}

type LogResponse struct {
    Logs      []LogEntry `json:"logs"`
    TotalCPU  float64    `json:"total_cpu"`
    NodeCount int        `json:"node_count"`
}

var nodeConfig = map[string]NodeInfo{
	"worker-ubuntu-01": {ContainerName: "worker-ubuntu-01"},
	// เพิ่ม node อื่นๆ ตรงนี้
}

// 1. เพิ่ม struct server เพื่อจัดการคำสั่ง
type server struct {
	proto.UnimplementedMasterControlServer
	mu       sync.Mutex
	commands map[string]chan string // เก็บ Channel ของคำสั่งแยกตาม NodeID
}

func cleanupOldLogs() {
    // 💡 ตั้งค่าให้ลบข้อมูลที่เก่ากว่า 1 ชั่วโมง
    query := `DELETE FROM worker_logs WHERE created_at < NOW() - INTERVAL '1 hour'`
    
    result, err := db.Exec(query)
    if err != nil {
        log.Println("❌ Error cleaning up old logs:", err)
        return
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected > 0 {
        log.Printf("🧹 Cleaned up %d old log entries.", rowsAffected)
    }
}

func scaleUpNode() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Error connecting to Docker:", err)
		return
	}

	// --- ส่วนคำสั่งสร้าง Container ---
	// config := &container.Config{
	//  // <--- แก้ชื่อ Image ตรงนี้
	// 	Cmd:   []string{"./worker"},
	// 	Env:   []string{"MASTER_IP=192.168.1.166"},
	// 	Tty: false, // <--- แก้ IP ตรงนี้
	// }
	// hostConfig := &container.HostConfig{
	// 	Resources: container.Resources{
	// 		CPUShares: 512,
	// 		Memory:    1024 * 1024 * 1024, // 1GB
	// 	},
	// }

	// createOptions := container{
	// 	Config:     config,
	// 	HostConfig: hostConfig,
	// }


	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			CPUShares: 512,
			Memory:    1024 * 1024 * 1024,
		},
		// 🛠️ 2. ระบุ Network ที่ต้องการให้ Container ใหม่เข้าไปอยู่
		NetworkMode: "supabase_default", // <--- แก้เป็นชื่อ Network ของคุณ
	}

	// networkConfig := &network.NetworkingConfig{
	// 	EndpointsConfig: map[string]*network.EndpointSettings{
	// 		"supabase_default": &network.EndpointSettings{
	// 			// สามารถกำหนด IP แบบ Static ได้ที่นี่ (ถ้าต้องการ)
	// 			// IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: "172.20.0.10"},
	// 		},
	// 	},
	// }


	containerName := fmt.Sprintf("worker-node-%d", time.Now().Unix())
	resp, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Cmd:   []string{"./worker"},
			Env:   []string{
			"MASTER_IP=siam-synapse-master", // 💡 แนะนำ: ใช้ชื่อ Container ของ Master แทน IP
			fmt.Sprintf("NODE_ID=%s", containerName), // ส่ง NodeID ไปให้ Worker
		},
		Tty: false,
		},
		HostConfig: hostConfig, // HostConfig
    	//NetworkConfig: networkConfig,  
		Image: "jkfastdevth/worker-node:latest",
	})

	// resp, err := cli.ContainerCreate(ctx, &createOptions)

	if err != nil {
		log.Println("Error creating container:", err)
		return
	}


	// --- สั่ง Start Container ---
	// 🛠️ แก้ไข: ใช้ container.StartOptions แทน types.ContainerStartOptions
	if _, err := cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		//panic(err)
		log.Println("Error starting container:", err)
	}

	// cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	log.Println("New Node created with ID:", resp.ID)
}

func checkAndScale(currentNodes int, totalCpuUsage float64) {
	if currentNodes == 0 {
		return
	}
	averageCpu := totalCpuUsage / float64(currentNodes)

	if averageCpu > 70.0 && currentNodes < 5 {
		log.Println("⚡ High Load detected! Scaling up...")
		go scaleUpNode()
	}
}

func (s *server) GetCommand(req *proto.NodeStatus, stream proto.MasterControl_GetCommandServer) error {
	log.Printf("📡 Worker %s is listening for commands...", req.NodeId)

	s.mu.Lock()
	if s.commands == nil {
		s.commands = make(map[string]chan string)
	}
	if _, ok := s.commands[req.NodeId]; !ok {
		s.commands[req.NodeId] = make(chan string, 1)
	}
	cmdChan := s.commands[req.NodeId]
	s.mu.Unlock()

	for cmd := range cmdChan {
		log.Printf("📤 Sending command '%s' to %s", cmd, req.NodeId)
		if err := stream.Send(&proto.Command{Command: cmd}); err != nil {
			return err
		}
	}
	return nil
}

func countActiveNodes() int {
	ctx := context.Background()
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        log.Println("Error connecting to Docker:", err)
        return 0
    }

    // 🛠️ แก้ไข: ใช้ container.ListOptions แบบปกติ (ไม่จำเป็นต้อง All: true)
	containers, err := cli.ContainerList(ctx, client.ContainerListOptions{})
    // containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
    if err != nil {
        log.Println("Error listing containers:", err)
        return 0
    }

    count := 0
    for _, c := range containers.Items {
        // --- เช็คชื่อ Container และ Status ---
        // 1. เช็คว่าชื่อ Container มีคำว่า "/worker-node"
        // 2. เช็คว่า Status เป็น "running"
        if contains(c.Names, "/worker-ubuntu-01") && c.State == "running" {
            count++
        }
    }
    return count
}

// func countActiveNodes() int {
// 	ctx := context.Background()
// 	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
// 	if err != nil {
// 		log.Println("Error connecting to Docker:", err)
// 		return 0
// 	}

// 	// 🛠️ แก้ไข: ใช้ container.ListOptions แทน types.ContainerListOptions
// 	// containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
// 	containers, err := cli.ContainerList(ctx, client.ContainerListOptions{})

// 	if err != nil {
// 		log.Println("Error listing containers:", err)
// 		return 0
// 	}

// 	count := 0
// 	for _, c := range containers.Items {
// 		if contains(c.Names, "/worker-ubuntu-01") {
// 			count++
// 		}
// 	}
// 	return count
// }

 

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
func calculateTotalCpu() float64 {
    statsMu.RLock()
    defer statsMu.RUnlock()
    
    total := 0.0
    for _, cpu := range workerCpuMap {
        total += cpu
    }
    return total
}
func (s *server) ReportStatus(ctx context.Context, in *proto.NodeStatus) (*proto.Ack, error) {
	fmt.Printf("📡 Heartbeat from: %s\n", in.NodeId)

	query := `INSERT INTO worker_logs (node_id, cpu_usage, ram_usage, status) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, in.NodeId, in.CpuUsage, in.RamUsage, in.Status)

	if err != nil {
		fmt.Printf("❌ DB Error: %v\n", err)
		return &proto.Ack{Success: false, Message: "DB Error"}, nil
	}

	statsMu.Lock()
    workerCpuMap[in.NodeId] = float64(in.CpuUsage)
    statsMu.Unlock()

	totalCpu := calculateTotalCpu()
	nodeCount := countActiveNodes()
	checkAndScale(nodeCount, totalCpu)

	if in.CpuUsage > 90.0 {
		msg := fmt.Sprintf("⚠️ ALERT: High CPU Usage!\nNode: %s\nCPU: %.2f%%", in.NodeId, in.CpuUsage)
		go helper.SendLineAdminPush(msg)
	}

	return &proto.Ack{Success: true, Message: "Report Received"}, nil
}





func main() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	dsn := os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
        for {
            cleanupOldLogs()
            time.Sleep(1 * time.Hour)
        }
    }()

	masterServer := &server{}
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatal(err)
		}
		s := grpc.NewServer()
		proto.RegisterMasterControlServer(s, masterServer)
		log.Println("🚀 Master gRPC Server: Running on :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	app := fiber.New()
	app.Static("/", "../dashboard.html")

	app.Use("/ws", func(c *fiber.Ctx) error {
        if websocket.IsWebSocketUpgrade(c) {
            c.Locals("allowed", true)
            return c.Next()
        }
        return fiber.ErrUpgradeRequired
    })
	
	// 🛠️ 2. จัดการ WebSocket Connection
    app.Get("/ws/stats", websocket.New(func(c *websocket.Conn) {
        // ดึงข้อมูล stats ทุกๆ 1 วินาที
        for {
            statsMu.RLock()
            // สร้าง data JSON ที่จะส่ง

	rows, err := db.Query("SELECT node_id, cpu_usage, ram_usage, status, created_at FROM worker_logs ORDER BY created_at DESC LIMIT 10")
		if err != nil {
			log.Println("Error querying database:", err)
			return
		}
		defer rows.Close()
		//totalCpu := calculateTotalCpu() // ⚠️ ต้องเขียน logic จริง
    	//nodeCount := countActiveNodes()

	

		var logs []LogEntry
		for rows.Next() {
			var entry LogEntry
			if err := rows.Scan(&entry.NodeID, &entry.CPUUsage, &entry.RAMUsage, &entry.Status, &entry.CreatedAt); err != nil {
				log.Println("Error scanning row:", err)
				return
			}
			logs = append(logs, entry)
		}

		 
      

            data := fiber.Map{
				"logs":       logs,
                "total_cpu":  calculateTotalCpu(),
                "node_count": countActiveNodes(),
            }
            statsMu.RUnlock()

            if err := c.WriteJSON(data); err != nil {
                break // connection หลุด
            }
            time.Sleep(1 * time.Second) // ส่งข้อมูลทุก 1 วินาที
        }
    }))
	app.Post("/container/:nodeID/:action", func(c *fiber.Ctx) error {
		nodeID := c.Params("nodeID")
		action := c.Params("action")

		info, ok := nodeConfig[nodeID]
		if !ok {
			info = NodeInfo{ContainerName: nodeID}
		}

		masterServer.mu.Lock()
		cmdChan, ok := masterServer.commands[nodeID]
		masterServer.mu.Unlock()

		if !ok {
			return c.Status(404).JSON(fiber.Map{"message": "Worker node not connected (Stream)"})
		}
		cmdChan <- action

		log.Printf("Action '%s' sent via stream to %s (Container: %s)", action, nodeID, info.ContainerName)

		return c.JSON(fiber.Map{
			"message":   "Action queued",
			"node":      nodeID,
			"container": info.ContainerName,
			"action":    action,
		})
	})

	app.Get("/logs", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT node_id, cpu_usage, ram_usage, status, created_at FROM worker_logs ORDER BY created_at DESC LIMIT 10")
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()
		totalCpu := calculateTotalCpu() // ⚠️ ต้องเขียน logic จริง
    	nodeCount := countActiveNodes()

	

		var logs []LogEntry
		for rows.Next() {
			var entry LogEntry
			if err := rows.Scan(&entry.NodeID, &entry.CPUUsage, &entry.RAMUsage, &entry.Status, &entry.CreatedAt); err != nil {
				return c.Status(500).SendString(err.Error())
			}
			logs = append(logs, entry)
		}

		return c.JSON(fiber.Map{
        "logs":       logs,
        "total_cpu":  totalCpu,
        "node_count": nodeCount,
    })
		// return c.JSON(LogResponse{
        // Logs:      logs,
        // TotalCPU:  totalCpu,
        // NodeCount: nodeCount,
    	// })
	//	return c.JSON(logs)
	})

	log.Fatal(app.Listen(":8080"))
}