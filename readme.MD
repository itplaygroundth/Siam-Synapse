# OpenClaw Distributed AI Gateway

## 📋 ข้อมูลเบื้องต้น (Project Overview)
* **Project Name:** OpenClaw Distributed AI Gateway
* **Version:** 1.0 (Zero Day Edition)
* **Goal:** สร้างระบบ AI Gateway ที่ทนทาน (Resilient), ขยายตัวได้ (Scalable), และดูแลตัวเองได้ (Self-Healing)
* **Tech Stack:** Go (Fiber), gRPC, Redis, SQLite (Encrypted), Docker

---

## 🏗️ 1. สถาปัตยกรรมระบบ (Architecture)



### 1.1 Master Node (Orchestrator)
* **หน้าที่:** จัดการ Config/Keys (SQLite), รัน Dashboard (Fiber), รับ Heartbeat (gRPC Server)
* **Tech:** Fiber + gRPC Server + Encrypted SQLite

### 1.2 Worker Node (Proxy)
* **หน้าที่:** รัน Proxy หน้า Service (Fiber), ส่ง Heartbeat (gRPC Client), รับคำสั่ง (gRPC Client)
* **Tech:** Fiber + gRPC Client

---

## 🛠️ 2. Tech Stack & Library

| Component | Technology | Usage |
| :--- | :--- | :--- |
| **Language** | Go 1.22+ | Core Logic |
| **Framework** | **GoFiber** | HTTP API & Dashboard |
| **RPC** | **gRPC** | Communication between nodes |
| **Database** | SQLite + **SQLCipher** | Encrypted Config Storage |
| **Cache/Bus** | Redis | Heartbeat, Pub/Sub |
| **Deployment** | Docker & Compose | Containerization |

---

## 💻 3. ข้อกำหนดทางเทคนิค (Technical Specifications)

### 3.1 สภาพแวดล้อมปัจจุบัน (Current Host)
* **OS:** Ubuntu 22.04 LTS หรือใหม่กว่า
* **CPU:** Intel Core i3 Gen 12 (4 Cores / 8 Threads)
* **RAM:** 16GB
* **Storage:** 256GB SSD NVMe.2
* **Network:** Gigabit Ethernet (สำหรับติดต่อ Node ภายนอก)

### 3.2 ความต้องการขั้นต่ำสำหรับ Node อื่นๆ (Minimal Node Requirement)
* **OS:** Linux (Ubuntu/Debian)
* **CPU:** 1 Core
* **RAM:** 1GB
* **Storage:** 10GB SSD

### 3.3 การเตรียมตัวย้ายขึ้น Cloud (Cloud VPS Consideration)
ในอนาคตเมื่อย้ายขึ้น VPS (เช่น DigitalOcean, AWS, GCP) ให้พิจารณา:
* **Master Node:** เลือก VPS ที่มี CPU/RAM ระดับกลาง (เช่น 2 vCPU, 4GB RAM) เพื่อรองรับ gRPC และ Dashboard
* **Worker Nodes:** เลือก VPS ขนาดเล็ก (1 vCPU, 1GB RAM) เพื่อประหยัดค่าใช้จ่าย
* **Network:** ตรวจสอบ Firewall ให้เปิดพอร์ต gRPC และ HTTP/HTTPS

---

## 📝 4. แผนการดำเนินการ (Implementation Roadmap)

### 🗓️ Phase 1: Foundation (Day 1)
- [ ] **Setup Environment:** ติดตั้ง Docker, Go, และ Node.js บน Ubuntu
- [ ] **Git Setup:** สร้าง Repository, ตั้งค่า `.gitignore` และ Branching Strategy
- [ ] **Initialize Master Container:** สร้าง Docker Compose สำหรับ Redis และ SQLite
- [ ] **Encrypted Database:** พัฒนา Module สำหรับ Encrypt/Decrypt Config ด้วย SQLCipher

### 🗓️ Phase 2: Master Controller (Day 2-3)
- [ ] **gRPC Server:** พัฒนา gRPC Server บน Master Node สำหรับรับ Heartbeat จาก Worker
- [ ] **Config API:** พัฒนา API ด้วย Fiber สำหรับจัดการ API Keys และ Policy
- [ ] **Dashboard:** สร้าง UI พื้นฐานสำหรับดูสถานะ Node และจัดการ Keys

### 🗓️ Phase 3: Worker Node & Proxy (Day 4-5)
- [ ] **Worker Docker:** สร้าง Dockerfile สำหรับ Worker Node
- [ ] **Fiber Proxy:** พัฒนา Proxy Handler ด้วย Fiber (แทนที่ `net/http` เดิม)
- [ ] **gRPC Client:** พัฒนา gRPC Client ใน Worker เพื่อส่ง Heartbeat และรับคำสั่งจาก Master
- [ ] **Key Management:** Worker ดึง Config จาก Master ผ่าน gRPC เมื่อเริ่มทำงาน

### 🗓️ Phase 4: Self-Healing & Automation (Day 6-7)
- [ ] **Heartbeat Monitor:** Master ตรวจสอบสถานะ Node ผ่าน Redis
- [ ] **Remote Command:** Master ส่งคำสั่งผ่าน gRPC/SSH เพื่อ Restart Worker ที่ตาย
- [ ] **Telegram Integration:** เชื่อมต่อ Telegram Bot เพื่อแจ้งเตือนเมื่อ Node ตายหรือเมื่อมีการ Switch Model

---

## 🖥️ 5. โครงสร้างไฟล์และ Git Workflow



```text
openclaw/
├── .git/
├── .gitignore         // ignore .env, db files, vendor
├── master/
│   ├── main.go        // Fiber Server + gRPC Server
│   ├── config.go      // SQLCipher Logic
│   └── dashboard/     // Web UI files
├── worker/
│   ├── main.go        // Fiber Proxy + gRPC Client
│   └── proxy.go       // Proxy Handler
├── proto/
│   └── control.proto  // gRPC Definitions
├── docker-compose.yml // Orchestration
└── .env.example       // Environment Variables