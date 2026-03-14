# Goth - Authentication Microservice

A production-standard authentication service built with Go, featuring both HTTP REST and gRPC APIs.

## Overview

Goth is a scalable authentication microservice that provides secure user registration, email verification, login/logout, and user management capabilities. It supports dual-protocol access through both REST APIs (HTTP) and gRPC, making it flexible for different client architectures.

## Architecture

Goth follows a modular architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────┐
│                  Goth Microservice                   │
├─────────────────────────────────────────────────────┤
│  Routes Layer (Auth Routes)                         │
│  Handlers Layer (Business Logic)                    │
│  Middlewares (Authentication, Authorization)        │
│  Models (User Data Structure)                       │
│  Database Access Layer (MongoDB)                    │
└─────────────────────────────────────────────────────┘
```

### Technology Stack

- **Framework**: Gin (HTTP), gRPC (RPC)
- **Database**: MongoDB
- **Authentication**: JWT Tokens
- **Password Hashing**: bcrypt
- **Gomail**: For email verification
- **Language**: Go 1.16+

---

## HTTP Server

The HTTP server runs on port **8080** and provides REST API endpoints for authentication operations.

### HTTP Server Architecture

```mermaid
graph TB
    subgraph Client["Client Layer"]
        REST["REST API Clients"]
    end
    
    subgraph HTTPServer["HTTP Server (Port 8080)"]
        Router["Gin Router"]
        Routes["Auth Routes"]
        
        subgraph Handlers["Handler Layer"]
            Signup["Signup Handler"]
            Login["Login Handler"]
            Verify["VerifyEmail Handler"]
            Logout["Logout Handler"]
            GetUsers["GetUsers Handler"]
        end
        
        subgraph Middleware["Middleware Layer"]
            AuthMW["Authentication Middleware"]
        end
    end
    
    subgraph Auth["Authentication Layer"]
        JWT["JWT Token<br/>Generator"]
        BCrypt["Password<br/>Hashing"]
    end
    
    subgraph Storage["Storage Layer"]
        MongoDB["MongoDB<br/>Collection: users"]
        Index["Auth Index"]
    end
    
    subgraph Email["Service Layer"]
        EmailSvc["Email Service<br/>Verification"]
    end
    
    REST -->|HTTP Request| Router
    Router --> Routes
    Routes --> Middleware
    Middleware -->|Protected Routes| AuthMW
    Routes -->|Public Routes| Handlers
    AuthMW -->|Authenticated Routes| Handlers
    
    Handlers --> JWT
    Handlers --> BCrypt
    Handlers --> EmailSvc
    Handlers --> MongoDB
    MongoDB --> Index
```

---

## gRPC Server

The gRPC server runs on port **50051** and provides high-performance RPC services using Protocol Buffers.

### gRPC Methods

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| Signup | `SignupRequest` | `SignupResponse` | Register user via gRPC |
| Login | `LoginRequest` | `LoginResponse` | Authenticate user via gRPC |
| VerifyEmail | `VerifyEmailRequest` | `VerifyEmailResponse` | Verify email via gRPC |
| Logout | `LogoutRequest` | `LogoutResponse` | Logout user via gRPC |
| StreamGetUsers | `UserListRequest` | `User` (stream) | Get all users as stream |

### gRPC Server Architecture

```mermaid
graph TB
    subgraph Client["gRPC Client Layer"]
        GRPCClients["Go/Java/Python/Node.js<br/>gRPC Clients"]
    end
    
    subgraph GRPCServer["gRPC Server (Port 50051)"]
        Listener["TCP Listener<br/>:50051"]
        Reflection["gRPC Reflection<br/>Service"]
        
        subgraph GRPCHandlers["gRPC Handlers"]
            GRPCSignup["Signup RPC"]
            GRPCLogin["Login RPC"]
            GRPCVerify["VerifyEmail RPC"]
            GRPCLogout["Logout RPC"]
            GRPCStream["StreamGetUsers RPC"]
        end
    end
    
    subgraph Proto["Protocol Buffers"]
        PB["Proto Definitions<br/>auth.proto"]
        Generated["Generated Code<br/>auth.pb.go<br/>auth_grpc.pb.go"]
    end
    
    subgraph Shared["Shared Services"]
        SharedJWT["JWT Token<br/>Generator"]
        SharedBCrypt["Password<br/>Hashing"]
        SharedEmail["Email Service"]
    end
    
    subgraph Storage2["Storage Layer"]
        MongoDB2["MongoDB<br/>Collection: users"]
    end
    
    GRPCClients -->|gRPC Call| Listener
    Listener --> GRPCHandlers
    Listener --> Reflection
    
    PB -->|Generates| Generated
    Generated -->|Uses| GRPCHandlers
    
    GRPCHandlers --> SharedJWT
    GRPCHandlers --> SharedBCrypt
    GRPCHandlers --> SharedEmail
    GRPCHandlers --> MongoDB2
```

---

## Project Structure

```
goth/
├── main.go                    # Entry point, initializes HTTP & gRPC servers
├── go.mod                     # Go module definition
├── README.md                  # This file
├── handlers/
│   └── auth.go               # HTTP request handlers
├── grpc/
│   └── auth_grpc_server.go   # gRPC service implementations
├── routes/
│   └── auth.go               # HTTP route definitions
├── middlewares/
│   └── auth.go               # Authentication middleware
├── models/
│   └── user.go               # User data model
├── helpers/
│   ├── env.go                # Environment variable loading
│   ├── token.go              # JWT token generation/validation
│   └── email.go              # Email sending utilities
├── db/
│   ├── connect.go            # MongoDB connection
│   └── auth_index.go         # Database indexes
└── proto/
    ├── auth.proto            # Protocol Buffer definitions
    ├── auth.pb.go            # Generated Protobuf code
    └── auth_grpc.pb.go       # Generated gRPC code
```

---

## Data Flow

### User Registration (HTTP)

```mermaid
sequenceDiagram
    participant Client
    participant HTTPServer
    participant Handler
    participant Auth
    participant MongoDB
    participant Email
    
    Client->>HTTPServer: POST /api/auth/signup
    HTTPServer->>Handler: Route to Signup Handler
    Handler->>Auth: Hash Password (bcrypt)
    Handler->>Auth: Generate JWT Token
    Handler->>MongoDB: Insert User Document
    Handler->>Email: Send Verification Email
    Handler->>Client: Return SignupResponse
```

### User Login (HTTP)

```mermaid
sequenceDiagram
    participant Client
    participant HTTPServer
    participant Handler
    participant MongoDB
    participant Auth
    
    Client->>HTTPServer: POST /api/auth/login
    HTTPServer->>Handler: Route to Login Handler
    Handler->>MongoDB: Find User by Email
    Handler->>Auth: Verify Password (bcrypt)
    Handler->>Auth: Generate JWT Token
    Handler->>Client: Return LoginResponse with Token
```

---