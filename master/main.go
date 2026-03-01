package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	_ "github.com/lib/pq" // Postgres driver
	"github.com/jkfastdevth/Siam-Synapse/proto"
	helper	"github.com/jkfastdevth/Siam-Synapse/master/helper"
)

var db *sql.DB

 type MasterApp struct {
	mu           sync.RWMutex
	workerClients map[string]proto.MasterControlClient // เก็บ Client ตาม NodeID
}

var appState = &MasterApp{
	workerClients: make(map[string]proto.MasterControlClient),
}
// type masterServer struct {
// 	proto.UnimplementedMasterControlServer
// } 
// 1. เพิ่ม struct server เพื่อจัดการคำสั่ง
type server struct {
	proto.UnimplementedMasterControlServer
	mu       sync.Mutex
	commands map[string]chan string // เก็บ Channel ของคำสั่งแยกตาม NodeID
}

// 2. Implement GetCommand
func (s *server) GetCommand(req *proto.NodeStatus, stream proto.MasterControl_GetCommandServer) error {
	log.Printf("📡 Worker %s is listening for commands...", req.NodeId)
	
	s.mu.Lock()
	if s.commands == nil {
		s.commands = make(map[string]chan string)
	}
	// สร้าง Channel ใหม่สำหรับ Node นี้ถ้ายังไม่มี
	if _, ok := s.commands[req.NodeId]; !ok {
		s.commands[req.NodeId] = make(chan string, 1)
	}
	cmdChan := s.commands[req.NodeId]
	s.mu.Unlock()

	// คอยส่งคำสั่งจาก Channel ไปให้ Worker
	for cmd := range cmdChan {
		log.Printf("📤 Sending command '%s' to %s", cmd, req.NodeId)
		if err := stream.Send(&proto.Command{Command: cmd}); err != nil {
			return err
		}
	}
	return nil
}
// 3. ย้าย ReportStatus มาอยู่ใต้ struct server ตัวเดียวกัน
func (s *server) ReportStatus(ctx context.Context, in *proto.NodeStatus) (*proto.Ack, error) {
	fmt.Printf("📡 Heartbeat from: %s\n", in.NodeId)

	// SQL สำหรับ Insert ข้อมูล
	query := `INSERT INTO worker_logs (node_id, cpu_usage, ram_usage, status) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, in.NodeId, in.CpuUsage, in.RamUsage, in.Status)
	
	if err != nil {
		fmt.Printf("❌ DB Error: %v\n", err)
		return &proto.Ack{Success: false, Message: "DB Error"}, nil
	}

	appState.mu.Lock()
    if _, ok := appState.workerClients[in.NodeId]; !ok {
        // ในกรณีนี้ Worker ต้องรับคำสั่ง gRPC ด้วย
        // โค้ดนี้สมมติว่า Master รู้อยู่แล้วว่า Worker อยู่ IP ไหน
        // เพื่อความง่ายในขั้นตอนแรก: ใช้ Docker Service Discovery
        conn, err := grpc.Dial(in.NodeId+":50052", grpc.WithInsecure()) // สมมติ port 50052
        if err == nil {
            appState.workerClients[in.NodeId] = proto.NewMasterControlClient(conn)
            log.Printf("🔌 Connected to worker: %s", in.NodeId)
        }
    }
    appState.mu.Unlock()


	if in.CpuUsage > 90.0 {
       msg := fmt.Sprintf("⚠️ ALERT: High CPU Usage!\nNode: %s\nCPU: %.2f%%", in.NodeId, in.CpuUsage)
    	go helper.SendLineAdminPush(msg)
    }

	return &proto.Ack{Success: true, Message: "Report Received"}, nil
}

func sendContainerAction(client proto.MasterControlClient, nodeID string, action string, containerName string) (*proto.Ack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	return client.ManageContainer(ctx, &proto.ContainerAction{
		NodeId:        nodeID,
		Action:        action,
		ContainerName: containerName,
	})
}

func main() {
	// 1. เชื่อมต่อ Supabase (ใช้ Environment Variable จาก docker-compose)
	var err error
	dsn := os.Getenv("DATABASE_URL") // format: "postgres://user:pass@host:port/db?sslmode=disable"
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	
	masterServer := &server{}
	// 2. รัน gRPC Server (เหมือนเดิม)
	go func() {
		lis, _ := net.Listen("tcp", ":50051")
		s := grpc.NewServer()
		//proto.RegisterMasterControlServer(s, &masterServer{})
		proto.RegisterMasterControlServer(s, masterServer)
		log.Println("🚀 Master gRPC Server: Running on :50051")
		s.Serve(lis)
	}()

	// 3. Fiber Dashboard
	app := fiber.New()
 	app.Static("/", "../dashboard.html")

	app.Post("/command/:nodeID", func(c *fiber.Ctx) error {
		nodeID := c.Params("nodeID")
		command := c.Query("cmd") // เช่น /command/worker-01?cmd=restart

		masterServer.mu.Lock()
		if cmdChan, ok := masterServer.commands[nodeID]; ok {
			cmdChan <- command
			masterServer.mu.Unlock()
			return c.SendString(fmt.Sprintf("Command '%s' sent to %s", command, nodeID))
		}
		masterServer.mu.Unlock()
		
		return c.Status(404).SendString("Node not found")
	})


	// เพิ่ม Fiber API Route
	app.Post("/container/:nodeID/:action/:containerName", func(c *fiber.Ctx) error {
		nodeID := c.Params("nodeID")
		action := c.Params("action")
	//	containerName := c.Params("containerName")

		// appState.mu.RLock()
		// client, ok := appState.workerClients[nodeID]
		// appState.mu.RUnlock()

		masterServer.mu.Lock()
    	cmdChan, ok := masterServer.commands[nodeID]
    	masterServer.mu.Unlock()

		if !ok {
			log.Printf("⚠️ Node %s not found in commands map", nodeID) // <--- ใส่ log ไว้เช็ค
			return c.Status(404).JSON(fiber.Map{"message": "Worker node not connected"})
    	}
		cmdChan <- action // ส่ง action (เช่น "restart") ผ่าน stream
		// ส่งคำสั่ง!
		// res, err := client.ManageContainer(context.Background(), &proto.ContainerAction{
		// 	NodeId:        nodeID,
		// 	Action:        action,
		// 	ContainerName: containerName,
		// })

		//log.Printf("Action sent to %s: %s", nodeID, action)
		
		// --- ปรับปรุงตรงนี้ ---
			// if err != nil {
			// 	log.Printf("❌ Error sending action to %s: %v", nodeID, err)
				
			// 	// ถ้า error คือ EOF หรือ Unavailable ให้ลบ client ออก
			// 	// เพื่อให้ ReportStatus สร้าง connection ใหม่ในครั้งถัดไป
			// 	appState.mu.Lock()
			// 	delete(appState.workerClients, nodeID)
			// 	appState.mu.Unlock()
				
			// 	return c.Status(500).JSON(fiber.Map{"message": "Worker is restarting or unavailable. Try again in a few seconds."})
			// }
			// ----------------------
		// if err != nil || !res.Success {
		// 	log.Printf("Error sending action to %s: %v", nodeID, err)
		// 	return c.Status(500).JSON(fiber.Map{"message": "Error message"})
		// }
		log.Printf("Action sent via stream to %s: %s", nodeID, action)
		
		return c.JSON(fiber.Map{"message": "Action sent to " + nodeID})
	})

	// แก้ไขตรงนี้: เพิ่ม Route /logs เพื่อดึงข้อมูลจาก DB
	app.Get("/logs", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT node_id, cpu_usage, ram_usage, status, created_at FROM worker_logs ORDER BY created_at DESC LIMIT 10")
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()

    type LogEntry struct {
        NodeID    string  `json:"node_id"`
        CPUUsage  float64 `json:"cpu_usage"`
        RAMUsage  float64 `json:"ram_usage"`
        Status    string  `json:"status"`
        CreatedAt string  `json:"created_at"`
    }

    var logs []LogEntry
    for rows.Next() {
        var entry LogEntry
        if err := rows.Scan(&entry.NodeID, &entry.CPUUsage, &entry.RAMUsage, &entry.Status, &entry.CreatedAt); err != nil {
            return c.Status(500).SendString(err.Error())
        }
        logs = append(logs, entry)
    }

    return c.JSON(logs)
})

	log.Fatal(app.Listen(":8080"))
}