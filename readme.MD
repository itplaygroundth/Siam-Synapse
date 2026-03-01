# Siam Synapse (Distributed AI Gateway)

## 📋 ข้อมูลเบื้องต้น (Project Overview)
* **Project Name:** Siam Synapse
* **Version:** 1.0 (Zero Day Edition)
* **Goal:** สร้างระบบ AI Gateway ที่ทนทาน (Resilient), ขยายตัวได้ (Scalable), และดูแลตัวเองได้ (Self-Healing)
* **Tech Stack:** Go (Fiber), gRPC, **Supabase (PostgreSQL)**, **Redis (Existing)**, Docker

---

## 🏗️ 1. สถาปัตยกรรมระบบ (Architecture)



### 1.1 Master Node (Orchestrator - Go)
* **หน้าที่:** จัดการ Config/Keys (ใช้ **PostgreSQL ของ Supabase**), รัน Dashboard (Fiber), รับ Heartbeat (gRPC Server)

### 1.2 Worker Node (Proxy - Go)
* **หน้าที่:** รัน Proxy หน้า Service (Fiber), ส่ง Heartbeat (gRPC Client), รับคำสั่ง (gRPC Client)
* **Tech:** Fiber + gRPC Client

---

## 🛠️ 2. Tech Stack & Library

| Component | Technology | Usage |
| :--- | :--- | :--- |
| **Language** | Go 1.22+ | Core Logic |
| **Framework** | **GoFiber** | HTTP API & Dashboard |
| **RPC** | **gRPC** | Communication between nodes |
| **Database** | **Supabase (PostgreSQL)** | Centralized Config Storage |
| **Cache/Bus** | **Redis (Existing)** | Heartbeat, Pub/Sub |
| **Deployment** | Docker & Compose | Containerization |

---

## 💻 3. ข้อกำหนดทางเทคนิค (Technical Specifications)

### 3.1 สภาพแวดล้อมปัจจุบัน (Current Host - Ubuntu)
* **OS:** Ubuntu 22.04 LTS
* **CPU:** Intel Core i3 Gen 12 (4 Cores / 8 Threads)
* **RAM:** 16GB (⚠️ **แน่น** - ต้องจัดการดีๆ)
* **Storage:** 256GB SSD NVMe.2
* **Installed Tech:** Supabase Stack (Postgres, Kong, etc.), Redis-Autobot

---

## 📝 4. แผนการดำเนินการ (Implementation Roadmap)

### 🗓️ Phase 1: Foundation & Integration (Day 1)
- [ ] **Git Setup:** CloneRepo `Siam-Synapse`
- [ ] **Database Setup:** สร้าง Table ใน Supabase สำหรับเก็บ API Keys/Config
- [ ] **Docker Compose:** สร้าง `docker-compose.yml` สำหรับรัน Master และ Worker โดยใช้ Redis เดิม

### 🗓️ Phase 2: Master Controller (Day 2-3)
- [ ] **gRPC Server:** พัฒนา gRPC Server บน Master Node
- [ ] **Config API:** พัฒนา API ด้วย Fiber เพื่อดึง/แก้ Config จาก Supabase
- [ ] **Dashboard:** สร้าง UI พื้นฐานเพื่อดูสถานะ Worker Node

### 🗓️ Phase 3: Worker Node & Proxy (Day 4-5)
- [ ] **Fiber Proxy:** พัฒนา Proxy Handler ด้วย Fiber
- [ ] **gRPC Client:** พัฒนา gRPC Client ใน Worker เพื่อส่ง Heartbeat
- [ ] **Key Management:** Worker ดึง Config จาก Master ผ่าน gRPC

---

## 🖥️ 5. โครงสร้างไฟล์ (File Structure)

```text
Siam-Synapse/
├── master/
│   ├── main.go        // Fiber Server + gRPC Server
│   ├── db.go          // Supabase Connect Logic
│   └── dashboard/
├── worker/
│   ├── main.go        // Fiber Proxy + gRPC Client
│   └── proxy.go
├── proto/
│   └── control.proto  // gRPC Definitions
├── docker-compose.yml // Orchestration (pointing to Redis-Autobot)
└── .gitignore