# Mini Agri Supply Chain Platform - Technical Specification

**Version**: 1.0.0  
**Last Updated**: March 2026  
**Target**: Portfolio project untuk BE

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Architecture](#2-architecture)
3. [Service Specifications](#3-service-specifications)
4. [API Specifications](#4-api-specifications)
5. [Database Schema](#5-database-schema)
6. [Integration Specifications](#6-integration-specifications)
7. [Infrastructure](#7-infrastructure)
8. [Security](#8-security)
9. [Testing Strategy](#9-testing-strategy)
10. [Monitoring & Observability](#10-monitoring--observability)
11. [Deployment](#11-deployment)
12. [Development Workflow](#12-development-workflow)

---

## 1. System Overview

### 1.1 Purpose
Mini platform untuk mengelola supply chain produk pertanian dari farmer hingga distributor, dengan fitur real-time tracking dan transaction management.

### 1.2 Scope
**In Scope:**
- User management (multi-role)
- Product catalog & inventory
- Order processing & transactions
- Real-time delivery tracking
- Analytics & reporting
- Third-party API integration

**Out of Scope (v1.0):**
- Payment gateway integration (simulated only)
- Mobile application
- ML-based recommendations
- Multi-language support

### 1.3 Target Metrics
```
Performance:
- API Response Time: p95 < 200ms, p99 < 500ms
- Throughput: 1000 req/s per service
- Availability: 99.9% uptime

Quality:
- Test Coverage: > 80%
- Code Quality: SonarQube grade A
- Zero critical security vulnerabilities
```

### 1.4 Technology Stack

**Backend:**
- Language: Go 1.21+
- Frameworks: Gin (REST), grpc-go (gRPC), gqlgen (GraphQL)
- ORM: GORM v2

**Databases:**
- Primary: PostgreSQL 15+
- Cache: Redis 7+
- Message Queue: RabbitMQ 3.12+

**Infrastructure:**
- Containerization: Docker 24+, Docker Compose
- Orchestration: Kubernetes 1.28+
- Cloud: GCP (primary) or AWS
- CI/CD: GitHub Actions
- IaC: Terraform

**Monitoring:**
- Metrics: Prometheus + Grafana
- Logging: ELK Stack (Elasticsearch, Logstash, Kibana)
- Tracing: Jaeger
- APM: OpenTelemetry

---

## 2. Architecture

### 2.1 High-Level Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                         Internet                             │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │   Load Balancer │
              │   (Nginx/GCP LB)│
              └────────┬─────────┘
                       │
                       ▼
              ┌─────────────────┐
              │   API Gateway   │◄──── Rate Limiting
              │   (REST/GraphQL)│◄──── Authentication
              └────────┬─────────┘
                       │
       ┌───────────────┼───────────────┐
       │               │               │
       ▼               ▼               ▼
┌──────────┐    ┌──────────┐    ┌──────────┐
│  User    │◄──►│ Product  │◄──►│  Order   │
│ Service  │    │ Service  │    │ Service  │
└────┬─────┘    └────┬─────┘    └────┬─────┘
     │               │               │
     │          ┌────▼─────┐         │
     │          │  Redis   │         │
     │          │  Cache   │         │
     │          └──────────┘         │
     │                               │
     ▼                               ▼
┌──────────┐                   ┌──────────┐
│Tracking  │                   │Analytics │
│Service   │◄──────────────────│Service   │
└────┬─────┘                   └────┬─────┘
     │                              │
     │         ┌──────────┐         │
     └────────►│RabbitMQ  │◄────────┘
               │Event Bus │
               └──────────┘
                    │
                    ▼
              ┌──────────┐
              │PostgreSQL│
              │ Primary  │
              └──────────┘
```

### 2.2 Service Communication Patterns

**Synchronous Communication:**
```
REST API:
- API Gateway → External clients
- Service-to-service (simple queries)

gRPC:
- Internal service-to-service (performance-critical)
- User Service ↔ Order Service
- Product Service ↔ Inventory checks
```

**Asynchronous Communication:**
```
RabbitMQ Events:
- Order Created → Inventory Update, Notification
- Delivery Status Changed → Analytics, Notification
- Stock Low → Alert Service

Exchanges:
- topic.orders (fanout pattern)
- topic.inventory (direct pattern)
- topic.notifications (topic pattern)
```

### 2.3 Data Flow Examples

**Example 1: Create Order Flow**
```
1. Client → API Gateway: POST /v1/orders
2. API Gateway → Order Service (gRPC): CreateOrder()
3. Order Service → User Service (gRPC): ValidateUser()
4. Order Service → Product Service (gRPC): CheckStock()
5. Order Service → PostgreSQL: INSERT INTO orders
6. Order Service → RabbitMQ: Publish(OrderCreated)
7. Product Service ← RabbitMQ: Consume(OrderCreated) → Update stock
8. Analytics Service ← RabbitMQ: Consume(OrderCreated) → Log event
9. Order Service → Client: 201 Created
```

**Example 2: Real-time Tracking Flow**
```
1. Driver App → API Gateway: WebSocket connection
2. API Gateway → Tracking Service: Upgrade to WebSocket
3. Tracking Service: Maintain connection pool
4. Driver sends GPS: {lat, lng, timestamp}
5. Tracking Service → PostgreSQL: INSERT location
6. Tracking Service → Redis: SET driver:123:location (TTL 60s)
7. Tracking Service → WebSocket: Broadcast to subscribers
8. Customer App ← WebSocket: Receive real-time location
```

---

## 3. Service Specifications

### 3.1 API Gateway

**Responsibilities:**
- Route incoming requests to appropriate services
- Authentication & authorization (JWT)
- Rate limiting (per user/IP)
- Request/response transformation
- API versioning

**Tech Stack:**
- Framework: Gin (Go)
- GraphQL: gqlgen
- Auth: JWT (golang-jwt/jwt)

**Endpoints:**
```
REST:
- /v1/auth/*
- /v1/users/*
- /v1/products/*
- /v1/orders/*
- /v1/tracking/*

GraphQL:
- /graphql (unified query endpoint)

WebSocket:
- /ws/tracking
```

**Configuration:**
```yaml
# config/gateway.yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

rate_limit:
  requests_per_minute: 100
  burst: 20

jwt:
  secret: ${JWT_SECRET}
  expiry: 24h
  refresh_expiry: 168h

cors:
  allowed_origins:
    - http://localhost:3000
    - https://app.example.com
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
```

### 3.2 User Service

**Responsibilities:**
- User registration & authentication
- Role-based access control (RBAC)
- User profile management
- Session management

**Database Tables:**
```sql
-- users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
```

**gRPC Service Definition:**
```protobuf
// proto/user/v1/user.proto
syntax = "proto3";

package user.v1;

option go_package = "github.com/yourusername/agri-platform/proto/user/v1";

service UserService {
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc ValidateUser(ValidateUserRequest) returns (ValidateUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}

message CreateUserRequest {
  string email = 1;
  string password = 2;
  string full_name = 3;
  string phone = 4;
  string role = 5;
}

message CreateUserResponse {
  string user_id = 1;
  string email = 2;
  string full_name = 3;
  string role = 4;
}

message ValidateUserRequest {
  string user_id = 1;
}

message ValidateUserResponse {
  bool is_valid = 1;
  string role = 2;
  string status = 3;
}
```

**Business Logic:**
```go
// internal/user/service.go
type Service struct {
    repo Repository
    hasher PasswordHasher
    logger *zap.Logger
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Validate email format
    if !isValidEmail(req.Email) {
        return nil, ErrInvalidEmail
    }
    
    // Check email uniqueness
    exists, err := s.repo.EmailExists(ctx, req.Email)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, ErrEmailAlreadyExists
    }
    
    // Hash password
    hash, err := s.hasher.Hash(req.Password)
    if err != nil {
        return nil, err
    }
    
    // Create user
    user := &User{
        Email:        req.Email,
        PasswordHash: hash,
        FullName:     req.FullName,
        Phone:        req.Phone,
        Role:         req.Role,
        Status:       "active",
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    s.logger.Info("user created", zap.String("user_id", user.ID))
    return user, nil
}
```

### 3.3 Product Service

**Responsibilities:**
- Product catalog management (CRUD)
- Inventory tracking
- Stock level monitoring
- Price management
- Cache management (Redis)

**Database Tables:**
```sql
-- products table
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    min_stock_threshold INTEGER DEFAULT 10,
    image_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_name ON products USING gin(to_tsvector('indonesian', name));

-- stock_movements table
CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id),
    movement_type VARCHAR(20) NOT NULL, -- 'in', 'out', 'adjustment'
    quantity INTEGER NOT NULL,
    reference_id UUID, -- order_id or other reference
    reference_type VARCHAR(50), -- 'order', 'return', 'adjustment'
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX idx_stock_movements_reference ON stock_movements(reference_id, reference_type);
```

**REST API Endpoints:**
```
GET    /v1/products              - List products (with pagination, filter)
GET    /v1/products/:id          - Get product detail
POST   /v1/products              - Create product
PUT    /v1/products/:id          - Update product
DELETE /v1/products/:id          - Soft delete product
GET    /v1/products/:id/stock    - Get stock movements
POST   /v1/products/:id/stock    - Adjust stock
```

**Cache Strategy:**
```go
// Cache key patterns
const (
    ProductCacheKey       = "product:%s"           // product:uuid
    ProductListCacheKey   = "products:list:%s"     // products:list:category:fertilizer
    ProductStockCacheKey  = "product:stock:%s"     // product:stock:uuid
)

// Cache TTL
const (
    ProductCacheTTL     = 1 * time.Hour
    ProductListCacheTTL = 5 * time.Minute
    ProductStockCacheTTL = 30 * time.Second  // Short TTL for stock
)

// Cache invalidation strategy
func (s *Service) UpdateProduct(ctx context.Context, id string, req *UpdateProductRequest) error {
    // Update database
    if err := s.repo.Update(ctx, id, req); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := fmt.Sprintf(ProductCacheKey, id)
    if err := s.cache.Delete(ctx, cacheKey); err != nil {
        s.logger.Warn("failed to invalidate cache", zap.Error(err))
    }
    
    // Invalidate list cache (all variations)
    pattern := "products:list:*"
    if err := s.cache.DeletePattern(ctx, pattern); err != nil {
        s.logger.Warn("failed to invalidate list cache", zap.Error(err))
    }
    
    return nil
}
```

**Stock Management Logic:**
```go
// Reserve stock (when order created)
func (s *Service) ReserveStock(ctx context.Context, productID string, quantity int) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Lock row for update
    var currentStock int
    err = tx.QueryRow(
        "SELECT stock_quantity FROM products WHERE id = $1 FOR UPDATE",
        productID,
    ).Scan(&currentStock)
    if err != nil {
        return err
    }
    
    // Check availability
    if currentStock < quantity {
        return ErrInsufficientStock
    }
    
    // Update stock
    _, err = tx.Exec(
        "UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2",
        quantity, productID,
    )
    if err != nil {
        return err
    }
    
    // Record movement
    _, err = tx.Exec(
        "INSERT INTO stock_movements (product_id, movement_type, quantity) VALUES ($1, 'out', $2)",
        productID, quantity,
    )
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

### 3.4 Order Service

**Responsibilities:**
- Order creation & management
- Order status workflow
- Payment status tracking
- Order history

**Database Tables:**
```sql
-- orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    payment_method VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_number ON orders(order_number);

-- order_items table
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(12,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- order_status_history table
CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id),
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    changed_by UUID,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
```

**State Machine (Order Status):**
```go
// Order status workflow
const (
    StatusPending    = "pending"
    StatusConfirmed  = "confirmed"
    StatusProcessing = "processing"
    StatusShipped    = "shipped"
    StatusDelivered  = "delivered"
    StatusCancelled  = "cancelled"
)

var allowedTransitions = map[string][]string{
    StatusPending:    {StatusConfirmed, StatusCancelled},
    StatusConfirmed:  {StatusProcessing, StatusCancelled},
    StatusProcessing: {StatusShipped, StatusCancelled},
    StatusShipped:    {StatusDelivered},
    StatusDelivered:  {}, // Terminal state
    StatusCancelled:  {}, // Terminal state
}

func (s *Service) UpdateOrderStatus(ctx context.Context, orderID, newStatus string) error {
    order, err := s.repo.GetByID(ctx, orderID)
    if err != nil {
        return err
    }
    
    // Validate transition
    allowed := allowedTransitions[order.Status]
    if !contains(allowed, newStatus) {
        return ErrInvalidStatusTransition
    }
    
    // Update status
    oldStatus := order.Status
    order.Status = newStatus
    
    if err := s.repo.Update(ctx, order); err != nil {
        return err
    }
    
    // Record history
    if err := s.recordStatusChange(ctx, orderID, oldStatus, newStatus); err != nil {
        s.logger.Warn("failed to record status history", zap.Error(err))
    }
    
    // Publish event
    event := &OrderStatusChangedEvent{
        OrderID:   orderID,
        OldStatus: oldStatus,
        NewStatus: newStatus,
        ChangedAt: time.Now(),
    }
    if err := s.publisher.Publish(ctx, "orders.status.changed", event); err != nil {
        s.logger.Error("failed to publish event", zap.Error(err))
    }
    
    return nil
}
```

**Event Publishing:**
```go
// Event schemas
type OrderCreatedEvent struct {
    OrderID     string                 `json:"order_id"`
    UserID      string                 `json:"user_id"`
    Items       []OrderItem            `json:"items"`
    TotalAmount float64                `json:"total_amount"`
    CreatedAt   time.Time              `json:"created_at"`
}

type OrderStatusChangedEvent struct {
    OrderID   string    `json:"order_id"`
    OldStatus string    `json:"old_status"`
    NewStatus string    `json:"new_status"`
    ChangedAt time.Time `json:"changed_at"`
}

// RabbitMQ publisher
type RabbitMQPublisher struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, routingKey string, event interface{}) error {
    body, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    return p.channel.PublishWithContext(
        ctx,
        "orders",     // exchange
        routingKey,   // routing key
        false,        // mandatory
        false,        // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            Timestamp:    time.Now(),
        },
    )
}
```

### 3.5 Tracking Service

**Responsibilities:**
- Real-time GPS location tracking
- Delivery assignment
- Route history
- WebSocket connections management

**Database Tables:**
```sql
-- deliveries table
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID UNIQUE NOT NULL,
    driver_id UUID NOT NULL,
    status VARCHAR(20) DEFAULT 'assigned',
    pickup_location GEOGRAPHY(POINT),
    dropoff_location GEOGRAPHY(POINT),
    estimated_delivery TIMESTAMP,
    actual_delivery TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_deliveries_order_id ON deliveries(order_id);
CREATE INDEX idx_deliveries_driver_id ON deliveries(driver_id);
CREATE INDEX idx_deliveries_status ON deliveries(status);

-- location_tracking table
CREATE TABLE location_tracking (
    id BIGSERIAL PRIMARY KEY,
    delivery_id UUID REFERENCES deliveries(id),
    driver_id UUID NOT NULL,
    location GEOGRAPHY(POINT) NOT NULL,
    speed DECIMAL(5,2),
    heading DECIMAL(5,2),
    accuracy DECIMAL(5,2),
    recorded_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_location_tracking_delivery_id ON location_tracking(delivery_id);
CREATE INDEX idx_location_tracking_recorded_at ON location_tracking(recorded_at DESC);
CREATE INDEX idx_location_tracking_location ON location_tracking USING GIST(location);

-- Partitioning by month (for performance)
CREATE TABLE location_tracking_2026_03 PARTITION OF location_tracking
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
```

**WebSocket Handler:**
```go
// WebSocket connection manager
type ConnectionManager struct {
    connections map[string]*websocket.Conn // driverID -> connection
    mu          sync.RWMutex
    upgrader    websocket.Upgrader
}

func (cm *ConnectionManager) HandleWebSocket(c *gin.Context) {
    driverID := c.Query("driver_id")
    
    conn, err := cm.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "upgrade failed"})
        return
    }
    
    cm.addConnection(driverID, conn)
    defer cm.removeConnection(driverID)
    
    for {
        var location LocationUpdate
        if err := conn.ReadJSON(&location); err != nil {
            break
        }
        
        // Process location update
        if err := cm.processLocationUpdate(driverID, &location); err != nil {
            log.Printf("error processing location: %v", err)
            continue
        }
        
        // Broadcast to subscribers
        cm.broadcastLocation(driverID, &location)
    }
}

func (cm *ConnectionManager) processLocationUpdate(driverID string, loc *LocationUpdate) error {
    ctx := context.Background()
    
    // Save to database
    tracking := &LocationTracking{
        DriverID:   driverID,
        DeliveryID: loc.DeliveryID,
        Location:   postgis.Point{Lat: loc.Latitude, Lng: loc.Longitude},
        Speed:      loc.Speed,
        Heading:    loc.Heading,
        Accuracy:   loc.Accuracy,
        RecordedAt: time.Now(),
    }
    
    if err := cm.repo.SaveLocation(ctx, tracking); err != nil {
        return err
    }
    
    // Update Redis cache (for quick lookup)
    cacheKey := fmt.Sprintf("driver:location:%s", driverID)
    data, _ := json.Marshal(loc)
    if err := cm.redis.Set(ctx, cacheKey, data, 60*time.Second).Err(); err != nil {
        log.Printf("redis error: %v", err)
    }
    
    // Check if near destination (geofencing)
    if cm.isNearDestination(tracking) {
        cm.notifyNearDestination(driverID, tracking.DeliveryID)
    }
    
    return nil
}

// Geofencing check
func (cm *ConnectionManager) isNearDestination(tracking *LocationTracking) bool {
    delivery, err := cm.deliveryRepo.GetByID(context.Background(), tracking.DeliveryID)
    if err != nil {
        return false
    }
    
    // Calculate distance using PostGIS
    distance := calculateDistance(tracking.Location, delivery.DropoffLocation)
    return distance < 500 // 500 meters threshold
}
```

### 3.6 Analytics Service

**Responsibilities:**
- Event consumption from RabbitMQ
- Data aggregation
- Report generation
- Metrics calculation

**Database Tables:**
```sql
-- analytics_events table (event sourcing)
CREATE TABLE analytics_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    occurred_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX idx_analytics_events_occurred_at ON analytics_events(occurred_at DESC);
CREATE INDEX idx_analytics_events_data ON analytics_events USING GIN(event_data);

-- sales_summary table (materialized view)
CREATE MATERIALIZED VIEW sales_summary AS
SELECT 
    DATE_TRUNC('day', created_at) as date,
    COUNT(*) as total_orders,
    SUM(total_amount) as total_revenue,
    AVG(total_amount) as avg_order_value,
    COUNT(DISTINCT user_id) as unique_customers
FROM orders
WHERE status = 'delivered'
GROUP BY DATE_TRUNC('day', created_at);

CREATE UNIQUE INDEX idx_sales_summary_date ON sales_summary(date);

-- Refresh materialized view periodically
-- (via cron job or scheduled task)
```

**Event Consumer:**
```go
// RabbitMQ consumer
type EventConsumer struct {
    conn     *amqp.Connection
    channel  *amqp.Channel
    handlers map[string]EventHandler
}

func (ec *EventConsumer) Start(ctx context.Context) error {
    queue, err := ec.channel.QueueDeclare(
        "analytics-queue", // name
        true,              // durable
        false,             // delete when unused
        false,             // exclusive
        false,             // no-wait
        nil,               // arguments
    )
    if err != nil {
        return err
    }
    
    // Bind to exchanges
    bindings := []struct {
        exchange   string
        routingKey string
    }{
        {"orders", "orders.created"},
        {"orders", "orders.status.changed"},
        {"products", "products.stock.low"},
    }
    
    for _, b := range bindings {
        if err := ec.channel.QueueBind(
            queue.Name,
            b.routingKey,
            b.exchange,
            false,
            nil,
        ); err != nil {
            return err
        }
    }
    
    msgs, err := ec.channel.Consume(
        queue.Name,
        "",    // consumer
        false, // auto-ack (manual ack for reliability)
        false, // exclusive
        false, // no-local
        false, // no-wait
        nil,   // args
    )
    if err != nil {
        return err
    }
    
    // Process messages
    for msg := range msgs {
        if err := ec.handleMessage(ctx, msg); err != nil {
            log.Printf("error handling message: %v", err)
            msg.Nack(false, true) // Requeue
        } else {
            msg.Ack(false)
        }
    }
    
    return nil
}

func (ec *EventConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
    // Parse event
    var baseEvent struct {
        Type string `json:"type"`
    }
    if err := json.Unmarshal(msg.Body, &baseEvent); err != nil {
        return err
    }
    
    // Find handler
    handler, exists := ec.handlers[baseEvent.Type]
    if !exists {
        log.Printf("no handler for event type: %s", baseEvent.Type)
        return nil
    }
    
    // Execute handler
    return handler.Handle(ctx, msg.Body)
}
```

**Report Generation:**
```go
type ReportService struct {
    db *sql.DB
}

// Generate sales report
func (rs *ReportService) GenerateSalesReport(ctx context.Context, req *SalesReportRequest) (*SalesReport, error) {
    query := `
        SELECT 
            DATE_TRUNC($1, created_at) as period,
            COUNT(*) as orders_count,
            SUM(total_amount) as revenue,
            AVG(total_amount) as avg_order_value
        FROM orders
        WHERE created_at BETWEEN $2 AND $3
            AND status = 'delivered'
        GROUP BY period
        ORDER BY period DESC
    `
    
    rows, err := rs.db.QueryContext(ctx, query, 
        req.Granularity, // 'day', 'week', 'month'
        req.StartDate,
        req.EndDate,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var report SalesReport
    for rows.Next() {
        var entry SalesReportEntry
        if err := rows.Scan(
            &entry.Period,
            &entry.OrdersCount,
            &entry.Revenue,
            &entry.AvgOrderValue,
        ); err != nil {
            return nil, err
        }
        report.Entries = append(report.Entries, entry)
    }
    
    // Calculate totals
    report.TotalOrders = sumInt(report.Entries, func(e SalesReportEntry) int { return e.OrdersCount })
    report.TotalRevenue = sumFloat(report.Entries, func(e SalesReportEntry) float64 { return e.Revenue })
    
    return &report, nil
}
```

---

## 4. API Specifications

### 4.1 REST API Standards

**Versioning:**
```
/v1/resource   - Current stable version
/v2/resource   - Next version (when breaking changes needed)
```

**HTTP Methods:**
```
GET    - Retrieve resource(s)
POST   - Create new resource
PUT    - Full update (replace entire resource)
PATCH  - Partial update (modify specific fields)
DELETE - Remove resource (soft delete preferred)
```

**Status Codes:**
```
200 OK                - Successful GET, PUT, PATCH
201 Created           - Successful POST
204 No Content        - Successful DELETE
400 Bad Request       - Invalid input
401 Unauthorized      - Missing/invalid auth
403 Forbidden         - Insufficient permissions
404 Not Found         - Resource doesn't exist
409 Conflict          - Duplicate/constraint violation
422 Unprocessable     - Validation error
429 Too Many Requests - Rate limit exceeded
500 Internal Error    - Server error
503 Service Unavailable - Temporary unavailable
```

**Response Format:**
```json
// Success response
{
  "data": {
    "id": "uuid",
    "name": "Product Name",
    "...": "..."
  },
  "meta": {
    "timestamp": "2026-03-26T10:00:00Z",
    "request_id": "req-123"
  }
}

// Error response
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  },
  "meta": {
    "timestamp": "2026-03-26T10:00:00Z",
    "request_id": "req-123"
  }
}

// Paginated list response
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  },
  "meta": {
    "timestamp": "2026-03-26T10:00:00Z",
    "request_id": "req-123"
  }
}
```

### 4.2 OpenAPI Specification Example
```yaml
openapi: 3.0.0
info:
  title: Agri Supply Chain API
  version: 1.0.0
  description: API for agricultural supply chain management
  contact:
    name: Your Name
    email: your.email@example.com

servers:
  - url: https://api.agri-platform.example.com/v1
    description: Production
  - url: https://staging-api.agri-platform.example.com/v1
    description: Staging
  - url: http://localhost:8080/v1
    description: Local development

security:
  - BearerAuth: []

paths:
  /products:
    get:
      summary: List all products
      tags:
        - Products
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
        - name: per_page
          in: query
          schema:
            type: integer
            default: 20
            maximum: 100
        - name: category
          in: query
          schema:
            type: string
            enum: [fertilizer, pesticide, seeds, tools]
        - name: status
          in: query
          schema:
            type: string
            enum: [active, inactive]
            default: active
        - name: search
          in: query
          description: Search by product name
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Product'
                  pagination:
                    $ref: '#/components/schemas/Pagination'
                  meta:
                    $ref: '#/components/schemas/Meta'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalError'
    
    post:
      summary: Create a new product
      tags:
        - Products
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateProductRequest'
      responses:
        '201':
          description: Product created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/Product'
                  meta:
                    $ref: '#/components/schemas/Meta'
        '400':
          $ref: '#/components/responses/BadRequest'
        '422':
          $ref: '#/components/responses/ValidationError'

  /products/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
          format: uuid
    
    get:
      summary: Get product by ID
      tags:
        - Products
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/Product'
                  meta:
                    $ref: '#/components/schemas/Meta'
        '404':
          $ref: '#/components/responses/NotFound'
    
    put:
      summary: Update product
      tags:
        - Products
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateProductRequest'
      responses:
        '200':
          description: Product updated successfully
        '404':
          $ref: '#/components/responses/NotFound'
    
    delete:
      summary: Delete product
      tags:
        - Products
      responses:
        '204':
          description: Product deleted successfully
        '404':
          $ref: '#/components/responses/NotFound'

  /orders:
    post:
      summary: Create a new order
      tags:
        - Orders
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '201':
          description: Order created successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    $ref: '#/components/schemas/Order'

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    Product:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        description:
          type: string
        category:
          type: string
          enum: [fertilizer, pesticide, seeds, tools]
        unit:
          type: string
          enum: [kg, liter, piece]
        price:
          type: number
          format: decimal
        stock_quantity:
          type: integer
        min_stock_threshold:
          type: integer
        image_url:
          type: string
          format: uri
        status:
          type: string
          enum: [active, inactive]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

    CreateProductRequest:
      type: object
      required:
        - name
        - category
        - unit
        - price
        - stock_quantity
      properties:
        name:
          type: string
          minLength: 3
          maxLength: 255
        description:
          type: string
          maxLength: 1000
        category:
          type: string
          enum: [fertilizer, pesticide, seeds, tools]
        unit:
          type: string
          enum: [kg, liter, piece]
        price:
          type: number
          format: decimal
          minimum: 0
        stock_quantity:
          type: integer
          minimum: 0
        min_stock_threshold:
          type: integer
          minimum: 0
          default: 10
        image_url:
          type: string
          format: uri

    Order:
      type: object
      properties:
        id:
          type: string
          format: uuid
        order_number:
          type: string
        user_id:
          type: string
          format: uuid
        items:
          type: array
          items:
            $ref: '#/components/schemas/OrderItem'
        total_amount:
          type: number
          format: decimal
        status:
          type: string
          enum: [pending, confirmed, processing, shipped, delivered, cancelled]
        payment_status:
          type: string
          enum: [unpaid, paid, refunded]
        created_at:
          type: string
          format: date-time

    OrderItem:
      type: object
      properties:
        product_id:
          type: string
          format: uuid
        product_name:
          type: string
        quantity:
          type: integer
        unit_price:
          type: number
          format: decimal
        subtotal:
          type: number
          format: decimal

    CreateOrderRequest:
      type: object
      required:
        - items
      properties:
        items:
          type: array
          minItems: 1
          items:
            type: object
            required:
              - product_id
              - quantity
            properties:
              product_id:
                type: string
                format: uuid
              quantity:
                type: integer
                minimum: 1
        notes:
          type: string
          maxLength: 500

    Pagination:
      type: object
      properties:
        page:
          type: integer
        per_page:
          type: integer
        total:
          type: integer
        total_pages:
          type: integer

    Meta:
      type: object
      properties:
        timestamp:
          type: string
          format: date-time
        request_id:
          type: string

    Error:
      type: object
      properties:
        code:
          type: string
        message:
          type: string
        details:
          type: array
          items:
            type: object
            properties:
              field:
                type: string
              message:
                type: string

  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                $ref: '#/components/schemas/Error'

    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                type: object
                properties:
                  code:
                    type: string
                    example: UNAUTHORIZED
                  message:
                    type: string
                    example: Missing or invalid authentication token

    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                type: object
                properties:
                  code:
                    type: string
                    example: NOT_FOUND
                  message:
                    type: string
                    example: Resource not found

    ValidationError:
      description: Validation error
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                $ref: '#/components/schemas/Error'

    InternalError:
      description: Internal server error
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                type: object
                properties:
                  code:
                    type: string
                    example: INTERNAL_ERROR
                  message:
                    type: string
                    example: An unexpected error occurred
```

### 4.3 GraphQL Schema
```graphql
# schema.graphql

type Query {
  # User queries
  me: User!
  user(id: ID!): User
  
  # Product queries
  products(
    page: Int = 1
    perPage: Int = 20
    category: ProductCategory
    search: String
  ): ProductConnection!
  product(id: ID!): Product
  
  # Order queries
  orders(
    page: Int = 1
    perPage: Int = 20
    status: OrderStatus
  ): OrderConnection!
  order(id: ID!): Order
  
  # Analytics queries
  salesReport(
    startDate: DateTime!
    endDate: DateTime!
    granularity: ReportGranularity!
  ): SalesReport!
}

type Mutation {
  # Auth mutations
  login(email: String!, password: String!): AuthPayload!
  register(input: RegisterInput!): AuthPayload!
  
  # Product mutations
  createProduct(input: CreateProductInput!): Product!
  updateProduct(id: ID!, input: UpdateProductInput!): Product!
  deleteProduct(id: ID!): Boolean!
  
  # Order mutations
  createOrder(input: CreateOrderInput!): Order!
  updateOrderStatus(id: ID!, status: OrderStatus!): Order!
  cancelOrder(id: ID!): Order!
}

type Subscription {
  # Real-time tracking
  deliveryLocationUpdated(deliveryId: ID!): LocationUpdate!
  
  # Order updates
  orderStatusChanged(orderId: ID!): OrderStatusUpdate!
}

# Types
type User {
  id: ID!
  email: String!
  fullName: String!
  phone: String
  role: UserRole!
  status: UserStatus!
  createdAt: DateTime!
}

enum UserRole {
  FARMER
  DISTRIBUTOR
  DRIVER
  ADMIN
}

enum UserStatus {
  ACTIVE
  INACTIVE
  SUSPENDED
}

type Product {
  id: ID!
  name: String!
  description: String
  category: ProductCategory!
  unit: ProductUnit!
  price: Decimal!
  stockQuantity: Int!
  minStockThreshold: Int!
  imageUrl: String
  status: ProductStatus!
  createdAt: DateTime!
  updatedAt: DateTime!
}

enum ProductCategory {
  FERTILIZER
  PESTICIDE
  SEEDS
  TOOLS
}

enum ProductUnit {
  KG
  LITER
  PIECE
}

enum ProductStatus {
  ACTIVE
  INACTIVE
}

type Order {
  id: ID!
  orderNumber: String!
  user: User!
  items: [OrderItem!]!
  totalAmount: Decimal!
  status: OrderStatus!
  paymentStatus: PaymentStatus!
  notes: String
  createdAt: DateTime!
  updatedAt: DateTime!
}

type OrderItem {
  id: ID!
  product: Product!
  quantity: Int!
  unitPrice: Decimal!
  subtotal: Decimal!
}

enum OrderStatus {
  PENDING
  CONFIRMED
  PROCESSING
  SHIPPED
  DELIVERED
  CANCELLED
}

enum PaymentStatus {
  UNPAID
  PAID
  REFUNDED
}

# Pagination
type ProductConnection {
  edges: [ProductEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type ProductEdge {
  node: Product!
  cursor: String!
}

type OrderConnection {
  edges: [OrderEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type OrderEdge {
  node: Order!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

# Input types
input RegisterInput {
  email: String!
  password: String!
  fullName: String!
  phone: String
  role: UserRole!
}

input CreateProductInput {
  name: String!
  description: String
  category: ProductCategory!
  unit: ProductUnit!
  price: Decimal!
  stockQuantity: Int!
  minStockThreshold: Int
  imageUrl: String
}

input UpdateProductInput {
  name: String
  description: String
  price: Decimal
  stockQuantity: Int
  minStockThreshold: Int
  imageUrl: String
  status: ProductStatus
}

input CreateOrderInput {
  items: [OrderItemInput!]!
  notes: String
}

input OrderItemInput {
  productId: ID!
  quantity: Int!
}

# Auth
type AuthPayload {
  token: String!
  refreshToken: String!
  user: User!
}

# Analytics
type SalesReport {
  entries: [SalesReportEntry!]!
  totalOrders: Int!
  totalRevenue: Decimal!
  averageOrderValue: Decimal!
}

type SalesReportEntry {
  period: DateTime!
  ordersCount: Int!
  revenue: Decimal!
  avgOrderValue: Decimal!
}

enum ReportGranularity {
  DAY
  WEEK
  MONTH
}

# Subscriptions
type LocationUpdate {
  deliveryId: ID!
  driverId: ID!
  latitude: Float!
  longitude: Float!
  speed: Float
  heading: Float
  timestamp: DateTime!
}

type OrderStatusUpdate {
  orderId: ID!
  oldStatus: OrderStatus!
  newStatus: OrderStatus!
  timestamp: DateTime!
}

# Scalars
scalar DateTime
scalar Decimal
```

---

## 5. Database Schema

### 5.1 Complete Schema Diagram
```
┌─────────────────┐
│     users       │
├─────────────────┤
│ id (PK)         │
│ email (UNIQUE)  │
│ password_hash   │
│ full_name       │
│ phone           │
│ role            │
│ status          │
│ created_at      │
│ updated_at      │
└────────┬────────┘
         │
         │ 1:N
         │
┌────────▼────────┐
│    orders       │
├─────────────────┤
│ id (PK)         │
│ order_number    │
│ user_id (FK)    │───┐
│ total_amount    │   │
│ status          │   │ 1:N
│ payment_status  │   │
│ created_at      │   │
└────────┬────────┘   │
         │            │
         │ 1:N        │
         │       ┌────▼────────────┐
┌────────▼──────┐│  order_items   │
│deliveries     ││├────────────────┤
├───────────────┤││ id (PK)        │
│id (PK)        │││ order_id (FK)  │
│order_id (FK)  │││ product_id (FK)│──┐
│driver_id (FK) │││ product_name   │  │
│status         │││ quantity       │  │
│pickup_loc     │││ unit_price     │  │
│dropoff_loc    ││└────────────────┘  │
└───────┬───────┘│                    │
        │        │                    │
        │ 1:N    │                    │
        │        │              ┌─────▼──────────┐
┌───────▼───────────┐           │   products     │
│location_tracking  │           ├────────────────┤
├───────────────────┤           │ id (PK)        │
│id (PK)            │           │ name           │
│delivery_id (FK)   │           │ description    │
│driver_id (FK)     │           │ category       │
│location (GEOGRAPHY)│          │ unit           │
│speed              │           │ price          │
│heading            │           │ stock_quantity │
│accuracy           │           │ status         │
│recorded_at        │           │ created_at     │
└───────────────────┘           └────────┬───────┘
                                         │
                                         │ 1:N
                                         │
                                ┌────────▼───────────┐
                                │stock_movements     │
                                ├────────────────────┤
                                │ id (PK)            │
                                │ product_id (FK)    │
                                │ movement_type      │
                                │ quantity           │
                                │ reference_id       │
                                │ reference_type     │
                                │ created_at         │
                                └────────────────────┘
```

### 5.2 Database Indexes Strategy
```sql
-- Performance-critical indexes

-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status) WHERE status = 'active';

-- Products
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_status ON products(status) WHERE status = 'active';
CREATE INDEX idx_products_stock ON products(stock_quantity) WHERE stock_quantity < min_stock_threshold;
CREATE INDEX idx_products_name_search ON products USING gin(to_tsvector('indonesian', name));

-- Orders (most queried table)
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_user_status ON orders(user_id, status); -- Composite for common query
CREATE INDEX idx_orders_number ON orders(order_number); -- Lookup by order number

-- Order Items
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- Deliveries
CREATE INDEX idx_deliveries_order_id ON deliveries(order_id);
CREATE INDEX idx_deliveries_driver_id ON deliveries(driver_id);
CREATE INDEX idx_deliveries_status ON deliveries(status);
CREATE INDEX idx_deliveries_pickup_loc ON deliveries USING GIST(pickup_location);
CREATE INDEX idx_deliveries_dropoff_loc ON deliveries USING GIST(dropoff_location);

-- Location Tracking (high-volume table)
CREATE INDEX idx_location_tracking_delivery_id ON location_tracking(delivery_id);
CREATE INDEX idx_location_tracking_driver_id ON location_tracking(driver_id);
CREATE INDEX idx_location_tracking_recorded_at ON location_tracking(recorded_at DESC);
CREATE INDEX idx_location_tracking_location ON location_tracking USING GIST(location);

-- Stock Movements
CREATE INDEX idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX idx_stock_movements_reference ON stock_movements(reference_id, reference_type);
CREATE INDEX idx_stock_movements_created_at ON stock_movements(created_at DESC);
```

### 5.3 Database Migrations
```sql
-- migrations/000001_initial_schema.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    role VARCHAR(50) NOT NULL CHECK (role IN ('farmer', 'distributor', 'driver', 'admin')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- ... (other tables)

-- migrations/000001_initial_schema.down.sql
DROP TABLE IF EXISTS location_tracking;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
```

**Migration Tool:**
```bash
# Using golang-migrate
migrate create -ext sql -dir db/migrations -seq create_users_table

# Apply migrations
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" up

# Rollback
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" down 1
```

---

## 6. Integration Specifications

### 6.1 Third-Party Integrations

**Weather API Integration (Example)**
```go
// internal/integrations/weather/client.go
type OpenWeatherClient struct {
    apiKey     string
    baseURL    string
    httpClient *http.Client
}

func NewOpenWeatherClient(apiKey string) *OpenWeatherClient {
    return &OpenWeatherClient{
        apiKey:  apiKey,
        baseURL: "https://api.openweathermap.org/data/2.5",
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *OpenWeatherClient) GetWeather(ctx context.Context, lat, lon float64) (*WeatherData, error) {
    url := fmt.Sprintf("%s/weather?lat=%f&lon=%f&appid=%s&units=metric",
        c.baseURL, lat, lon, c.apiKey)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("weather API error: %d", resp.StatusCode)
    }
    
    var weather WeatherData
    if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
        return nil, err
    }
    
    return &weather, nil
}
```

**Payment Gateway Integration (Simulated)**
```go
// internal/integrations/payment/midtrans.go
type MidtransClient struct {
    serverKey string
    baseURL   string
}

// For portfolio: just simulate the payment
func (c *MidtransClient) CreateTransaction(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    // In real implementation, call Midtrans API
    // For demo: simulate successful payment
    
    return &PaymentResponse{
        TransactionID: uuid.New().String(),
        Status:        "pending",
        PaymentURL:    "https://simulator.midtrans.com/payment",
        CreatedAt:     time.Now(),
    }, nil
}

func (c *MidtransClient) CheckStatus(ctx context.Context, transactionID string) (*PaymentStatus, error) {
    // Simulate payment status check
    return &PaymentStatus{
        TransactionID: transactionID,
        Status:        "success",
        PaidAt:        time.Now(),
    }, nil
}
```

### 6.2 Webhook Handling
```go
// internal/webhook/handler.go
type WebhookHandler struct {
    secretKey string
    handlers  map[string]WebhookProcessor
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
    // Verify signature
    signature := c.GetHeader("X-Webhook-Signature")
    body, _ := ioutil.ReadAll(c.Request.Body)
    
    if !h.verifySignature(body, signature) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
        return
    }
    
    // Parse webhook
    var webhook Webhook
    if err := json.Unmarshal(body, &webhook); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
        return
    }
    
    // Process webhook asynchronously
    go h.processWebhook(webhook)
    
    c.JSON(http.StatusOK, gin.H{"status": "accepted"})
}

func (h *WebhookHandler) verifySignature(body []byte, signature string) bool {
    mac := hmac.New(sha256.New, []byte(h.secretKey))
    mac.Write(body)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

---

## 7. Infrastructure

### 7.1 Docker Setup

**Dockerfile (Go service)**
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./main"]
```

**docker-compose.yml**
```yaml
version: '3.8'

services:
  # PostgreSQL
  postgres:
    image: postgis/postgis:15-3.3
    environment:
      POSTGRES_USER: agriuser
      POSTGRES_PASSWORD: agripass
      POSTGRES_DB: agri_platform
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U agriuser"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  # RabbitMQ
  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: agriuser
      RABBITMQ_DEFAULT_PASS: agripass
    ports:
      - "5672:5672"   # AMQP
      - "15672:15672" # Management UI
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5

  # API Gateway
  api-gateway:
    build:
      context: ./services/api-gateway
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=agriuser
      - DB_PASSWORD=agripass
      - DB_NAME=agri_platform
      - REDIS_HOST=redis:6379
      - RABBITMQ_URL=amqp://agriuser:agripass@rabbitmq:5672/
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
    restart: unless-stopped

  # User Service
  user-service:
    build:
      context: ./services/user-service
      dockerfile: Dockerfile
    ports:
      - "50051:50051" # gRPC
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=agriuser
      - DB_PASSWORD=agripass
      - DB_NAME=agri_platform
    depends_on:
      - postgres
    restart: unless-stopped

  # Product Service
  product-service:
    build:
      context: ./services/product-service
      dockerfile: Dockerfile
    ports:
      - "50052:50052" # gRPC
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  # Order Service
  order-service:
    build:
      context: ./services/order-service
      dockerfile: Dockerfile
    ports:
      - "50053:50053" # gRPC
    environment:
      - DB_HOST=postgres
      - RABBITMQ_URL=amqp://agriuser:agripass@rabbitmq:5672/
    depends_on:
      - postgres
      - rabbitmq
    restart: unless-stopped

  # Tracking Service
  tracking-service:
    build:
      context: ./services/tracking-service
      dockerfile: Dockerfile
    ports:
      - "8081:8081" # WebSocket
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  # Analytics Service
  analytics-service:
    build:
      context: ./services/analytics-service
      dockerfile: Dockerfile
    environment:
      - DB_HOST=postgres
      - RABBITMQ_URL=amqp://agriuser:agripass@rabbitmq:5672/
    depends_on:
      - postgres
      - rabbitmq
    restart: unless-stopped

  # Prometheus
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./infrastructure/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    restart: unless-stopped

  # Grafana
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./infrastructure/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./infrastructure/grafana/datasources:/etc/grafana/provisioning/datasources
    depends_on:
      - prometheus
    restart: unless-stopped

  # Jaeger (Distributed Tracing)
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "5775:5775/udp"
      - "6831:6831/udp"
      - "6832:6832/udp"
      - "5778:5778"
      - "16686:16686" # UI
      - "14268:14268"
      - "14250:14250"
      - "9411:9411"
    environment:
      - COLLECTOR_ZIPKIN_HOST_PORT=:9411
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:
  prometheus_data:
  grafana_data:
```

### 7.2 Kubernetes Manifests

**Namespace**
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: agri-platform
```

**ConfigMap**
```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: agri-platform
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "agri_platform"
  REDIS_HOST: "redis-service:6379"
  RABBITMQ_URL: "amqp://agriuser:agripass@rabbitmq-service:5672/"
```

**Secret**
```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: agri-platform
type: Opaque
stringData:
  DB_USER: agriuser
  DB_PASSWORD: agripass
  JWT_SECRET: your-super-secret-jwt-key-change-in-production
```

**Deployment (API Gateway)**
```yaml
# k8s/deployments/api-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: agri-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: your-registry/api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: DB_HOST
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: DB_USER
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: DB_PASSWORD
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: JWT_SECRET
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway-service
  namespace: agri-platform
spec:
  selector:
    app: api-gateway
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

**Horizontal Pod Autoscaler**
```yaml
# k8s/hpa/api-gateway-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
  namespace: agri-platform
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 8. Security

### 8.1 Authentication & Authorization

**JWT Implementation**
```go
// pkg/auth/jwt.go
type JWTManager struct {
    secretKey     string
    tokenDuration time.Duration
}

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func (m *JWTManager) Generate(user *User) (string, error) {
    claims := &Claims{
        UserID: user.ID,
        Email:  user.Email,
        Role:   user.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "agri-platform",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) Verify(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            return []byte(m.secretKey), nil
        },
    )
    
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    
    return claims, nil
}
```

**Middleware**
```go
// pkg/middleware/auth.go
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "missing authorization header",
            })
            return
        }
        
        // Extract token (Bearer <token>)
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid authorization header format",
            })
            return
        }
        
        claims, err := jwtManager.Verify(parts[1])
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid token",
            })
            return
        }
        
        // Store claims in context
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_role", claims.Role)
        
        c.Next()
    }
}

// Role-based authorization
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("user_role")
        if !exists {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "no role found",
            })
            return
        }
        
        role := userRole.(string)
        allowed := false
        for _, r := range roles {
            if r == role {
                allowed = true
                break
            }
        }
        
        if !allowed {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                "error": "insufficient permissions",
            })
            return
        }
        
        c.Next()
    }
}
```

### 8.2 Rate Limiting
```go
// pkg/middleware/ratelimit.go
type RateLimiter struct {
    redis   *redis.Client
    limit   int           // requests
    window  time.Duration // time window
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Use user ID if authenticated, otherwise IP
        key := c.ClientIP()
        if userID, exists := c.Get("user_id"); exists {
            key = userID.(string)
        }
        
        key = fmt.Sprintf("ratelimit:%s", key)
        
        // Increment counter
        count, err := rl.redis.Incr(c.Request.Context(), key).Result()
        if err != nil {
            c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "error": "rate limit check failed",
            })
            return
        }
        
        // Set expiry on first request
        if count == 1 {
            rl.redis.Expire(c.Request.Context(), key, rl.window)
        }
        
        // Check limit
        if count > int64(rl.limit) {
            c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
            c.Header("X-RateLimit-Remaining", "0")
            c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.window).Unix()))
            
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
            })
            return
        }
        
        // Set headers
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.limit-int(count)))
        
        c.Next()
    }
}
```

### 8.3 Input Validation
```go
// pkg/validator/validator.go
type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required,min=3,max=255"`
    Description string  `json:"description" binding:"max=1000"`
    Category    string  `json:"category" binding:"required,oneof=fertilizer pesticide seeds tools"`
    Unit        string  `json:"unit" binding:"required,oneof=kg liter piece"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    Stock       int     `json:"stock_quantity" binding:"required,gte=0"`
}

// Custom validators
func init() {
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("indonesian_phone", validateIndonesianPhone)
    }
}

func validateIndonesianPhone(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    re := regexp.MustCompile(`^(\+62|62|0)[0-9]{9,12}$`)
    return re.MatchString(phone)
}

// Usage in handler
func (h *Handler) CreateProduct(c *gin.Context) {
    var req CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "validation failed",
            "details": err.Error(),
        })
        return
    }
    
    // Process request
}
```

---

## 9. Testing Strategy

### 9.1 Unit Testing
```go
// internal/product/service_test.go
package product

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock repository
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, product *Product) error {
    args := m.Called(ctx, product)
    return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Product, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Product), args.Error(1)
}

// Test cases
func TestCreateProduct(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateProductRequest
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid product",
            input: &CreateProductRequest{
                Name:     "NPK Fertilizer",
                Category: "fertilizer",
                Unit:     "kg",
                Price:    50000,
                Stock:    100,
            },
            wantErr: false,
        },
        {
            name: "invalid price",
            input: &CreateProductRequest{
                Name:     "Product",
                Category: "fertilizer",
                Unit:     "kg",
                Price:    -100,
                Stock:    100,
            },
            wantErr: true,
            errMsg:  "price must be positive",
        },
        {
            name: "invalid stock",
            input: &CreateProductRequest{
                Name:     "Product",
                Category: "fertilizer",
                Unit:     "kg",
                Price:    50000,
                Stock:    -10,
            },
            wantErr: true,
            errMsg:  "stock cannot be negative",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockRepo := new(MockRepository)
            service := NewService(mockRepo, nil)
            
            if !tt.wantErr {
                mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
            }
            
            // Execute
            _, err := service.CreateProduct(context.Background(), tt.input)
            
            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
                mockRepo.AssertExpectations(t)
            }
        })
    }
}

func TestReserveStock(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewService(mockRepo, nil)
    
    productID := "product-123"
    
    // Mock current stock
    mockRepo.On("GetByID", mock.Anything, productID).Return(&Product{
        ID:            productID,
        StockQuantity: 50,
    }, nil)
    
    mockRepo.On("UpdateStock", mock.Anything, productID, 30).Return(nil)
    
    // Test successful reservation
    err := service.ReserveStock(context.Background(), productID, 20)
    assert.NoError(t, err)
    
    // Test insufficient stock
    err = service.ReserveStock(context.Background(), productID, 60)
    assert.Error(t, err)
    assert.Equal(t, ErrInsufficientStock, err)
}
```

### 9.2 Integration Testing
```go
// tests/integration/order_test.go
// +build integration

package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/suite"
)

type OrderTestSuite struct {
    suite.Suite
    db          *sql.DB
    userService *user.Service
    productService *product.Service
    orderService *order.Service
}

func (suite *OrderTestSuite) SetupSuite() {
    // Setup test database
    suite.db = setupTestDB()
    
    // Initialize services
    suite.userService = user.NewService(user.NewRepository(suite.db))
    suite.productService = product.NewService(product.NewRepository(suite.db))
    suite.orderService = order.NewService(order.NewRepository(suite.db))
}

func (suite *OrderTestSuite) TearDownSuite() {
    suite.db.Close()
}

func (suite *OrderTestSuite) SetupTest() {
    // Clean database before each test
    cleanDatabase(suite.db)
}

func (suite *OrderTestSuite) TestCreateOrder_Success() {
    ctx := context.Background()
    
    // Create test user
    user, err := suite.userService.Create(ctx, &CreateUserRequest{
        Email:    "test@example.com",
        Password: "password123",
        FullName: "Test User",
        Role:     "farmer",
    })
    suite.NoError(err)
    
    // Create test product
    product, err := suite.productService.Create(ctx, &CreateProductRequest{
        Name:     "NPK Fertilizer",
        Category: "fertilizer",
        Unit:     "kg",
        Price:    50000,
        Stock:    100,
    })
    suite.NoError(err)
    
    // Create order
    order, err := suite.orderService.Create(ctx, &CreateOrderRequest{
        UserID: user.ID,
        Items: []OrderItemRequest{
            {
                ProductID: product.ID,
                Quantity:  5,
            },
        },
    })
    
    suite.NoError(err)
    suite.NotNil(order)
    suite.Equal(user.ID, order.UserID)
    suite.Equal(1, len(order.Items))
    suite.Equal(float64(250000), order.TotalAmount) // 5 * 50000
    
    // Verify stock was reduced
    updatedProduct, err := suite.productService.GetByID(ctx, product.ID)
    suite.NoError(err)
    suite.Equal(95, updatedProduct.StockQuantity) // 100 - 5
}

func (suite *OrderTestSuite) TestCreateOrder_InsufficientStock() {
    ctx := context.Background()
    
    user, _ := suite.userService.Create(ctx, &CreateUserRequest{
        Email: "test@example.com",
        Password: "password123",
        FullName: "Test User",
        Role: "farmer",
    })
    
    product, _ := suite.productService.Create(ctx, &CreateProductRequest{
        Name:     "NPK Fertilizer",
        Category: "fertilizer",
        Unit:     "kg",
        Price:    50000,
        Stock:    10, // Only 10 in stock
    })
    
    // Try to order more than available
    _, err := suite.orderService.Create(ctx, &CreateOrderRequest{
        UserID: user.ID,
        Items: []OrderItemRequest{
            {
                ProductID: product.ID,
                Quantity:  20, // Requesting 20
            },
        },
    })
    
    suite.Error(err)
    suite.Equal(ErrInsufficientStock, err)
}

func TestOrderSuite(t *testing.T) {
    suite.Run(t, new(OrderTestSuite))
}
```

### 9.3 Load Testing (k6)
```javascript
// tests/load/create_order.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 }, // Ramp up to 100 users
    { duration: '5m', target: 100 }, // Stay at 100 users
    { duration: '2m', target: 200 }, // Ramp up to 200 users
    { duration: '5m', target: 200 }, // Stay at 200 users
    { duration: '2m', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.01'],   // Error rate should be less than 1%
  },
};

const BASE_URL = 'http://localhost:8080/v1';
let authToken = '';

export function setup() {
  // Login to get auth token
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: 'test@example.com',
    password: 'password123',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  authToken = JSON.parse(loginRes.body).data.token;
  return { token: authToken };
}

export default function (data) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${data.token}`,
  };
  
  // Create order
  const orderPayload = JSON.stringify({
    items: [
      {
        product_id: 'product-uuid-1',
        quantity: Math.floor(Math.random() * 10) + 1,
      },
    ],
  });
  
  const res = http.post(`${BASE_URL}/orders`, orderPayload, { headers });
  
  check(res, {
    'status is 201': (r) => r.status === 201,
    'response has order_id': (r) => JSON.parse(r.body).data.id !== undefined,
  });
  
  sleep(1); // Wait 1 second between iterations
}
```

**Run load test:**
```bash
k6 run tests/load/create_order.js
```

---

## 10. Monitoring & Observability

### 10.1 Prometheus Metrics
```go
// pkg/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP metrics
    HttpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    HttpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
    
    // Database metrics
    DbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"query_type"},
    )
    
    DbConnectionsOpen = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_open",
            Help: "Number of open database connections",
        },
    )
    
    // Business metrics
    OrdersCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "orders_created_total",
            Help: "Total number of orders created",
        },
    )
    
    OrderValue = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "order_value",
            Help:    "Order value in IDR",
            Buckets: []float64{10000, 50000, 100000, 500000, 1000000, 5000000},
        },
    )
    
    StockLevelLow = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "product_stock_low",
            Help: "Products with low stock (below threshold)",
        },
        []string{"product_id", "product_name"},
    )
)

// Middleware to record HTTP metrics
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", c.Writer.Status())
        
        HttpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        HttpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

**Prometheus Configuration**
```yaml
# infrastructure/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'api-gateway'
    static_configs:
      - targets: ['api-gateway:8080']
    metrics_path: '/metrics'

  - job_name: 'user-service'
    static_configs:
      - targets: ['user-service:50051']

  - job_name: 'product-service'
    static_configs:
      - targets: ['product-service:50052']

  - job_name: 'order-service'
    static_configs:
      - targets: ['order-service:50053']

  - job_name: 'tracking-service'
    static_configs:
      - targets: ['tracking-service:8081']

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### 10.2 Structured Logging
```go
// pkg/logger/logger.go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func New(environment string) (*zap.Logger, error) {
    var config zap.Config
    
    if environment == "production" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    
    config.OutputPaths = []string{"stdout"}
    config.ErrorOutputPaths = []string{"stderr"}
    
    return config.Build()
}

// Usage
logger, _ := logger.New("development")
defer logger.Sync()

logger.Info("order created",
    zap.String("order_id", order.ID),
    zap.String("user_id", order.UserID),
    zap.Float64("total_amount", order.TotalAmount),
    zap.Duration("processing_time", processingTime),
)

logger.Error("failed to create order",
    zap.Error(err),
    zap.String("user_id", userID),
    zap.Any("request", req),
)
```

### 10.3 Distributed Tracing (Jaeger)
```go
// pkg/tracing/tracing.go
package tracing

import (
    "io"
    
    "github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
)

func InitTracer(serviceName string) (opentracing.Tracer, io.Closer, error) {
    cfg := &config.Configuration{
        ServiceName: serviceName,
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:           true,
            LocalAgentHostPort: "jaeger:6831",
        },
    }
    
    tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
    if err != nil {
        return nil, nil, err
    }
    
    opentracing.SetGlobalTracer(tracer)
    return tracer, closer, nil
}

// Middleware
func TracingMiddleware(tracer opentracing.Tracer) gin.HandlerFunc {
    return func(c *gin.Context) {
        span := tracer.StartSpan(c.Request.URL.Path)
        defer span.Finish()
        
        span.SetTag("http.method", c.Request.Method)
        span.SetTag("http.url", c.Request.URL.String())
        
        ctx := opentracing.ContextWithSpan(c.Request.Context(), span)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
        
        span.SetTag("http.status_code", c.Writer.Status())
    }
}
```

---

## 11. Deployment

### 11.1 CI/CD Pipeline (GitHub Actions)
```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  GO_VERSION: '1.21'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgis/postgis:15-3.3
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test -v -race -coverprofile=coverage.out ./...
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: testuser
        DB_PASSWORD: testpass
        DB_NAME: testdb
        REDIS_HOST: localhost:6379
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella
    
    - name: Run integration tests
      run: go test -v -tags=integration ./tests/integration/...
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: testuser
        DB_PASSWORD: testpass
        DB_NAME: testdb
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest

  build:
    name: Build
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Log in to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy-staging:
    name: Deploy to Staging
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    environment: staging
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up kubectl
      uses: azure/setup-kubectl@v3
      with:
        version: 'v1.28.0'
    
    - name: Configure kubectl
      run: |
        echo "${{ secrets.KUBECONFIG }}" | base64 -d > kubeconfig.yaml
        export KUBECONFIG=kubeconfig.yaml
    
    - name: Deploy to staging
      run: |
        kubectl apply -f k8s/staging/ -n agri-platform-staging
        kubectl rollout status deployment/api-gateway -n agri-platform-staging

  deploy-production:
    name: Deploy to Production
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Deploy to GCP
      uses: google-github-actions/deploy-cloudrun@v1
      with:
        service: agri-platform-api
        image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
        region: asia-southeast2
        credentials: ${{ secrets.GCP_CREDENTIALS }}
```

### 11.2 Terraform (Infrastructure as Code)
```hcl
# infrastructure/terraform/main.tf
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
  
  backend "gcs" {
    bucket = "agri-platform-terraform-state"
    prefix = "prod"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# VPC Network
resource "google_compute_network" "vpc" {
  name                    = "agri-platform-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "subnet" {
  name          = "agri-platform-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.vpc.id
}

# Cloud SQL (PostgreSQL)
resource "google_sql_database_instance" "postgres" {
  name             = "agri-platform-db"
  database_version = "POSTGRES_15"
  region           = var.region
  
  settings {
    tier = "db-custom-2-7680"
    
    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.vpc.id
    }
    
    backup_configuration {
      enabled    = true
      start_time = "02:00"
    }
    
    database_flags {
      name  = "max_connections"
      value = "100"
    }
  }
  
  deletion_protection = true
}

resource "google_sql_database" "database" {
  name     = "agri_platform"
  instance = google_sql_database_instance.postgres.name
}

# GKE Cluster
resource "google_container_cluster" "primary" {
  name     = "agri-platform-gke"
  location = var.region
  
  remove_default_node_pool = true
  initial_node_count       = 1
  
  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
}

resource "google_container_node_pool" "primary_nodes" {
  name       = "primary-node-pool"
  location   = var.region
  cluster    = google_container_cluster.primary.name
  node_count = 3
  
  autoscaling {
    min_node_count = 3
    max_node_count = 10
  }
  
  node_config {
    preemptible  = false
    machine_type = "e2-standard-4"
    
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
}

# Redis (Memorystore)
resource "google_redis_instance" "cache" {
  name           = "agri-platform-redis"
  tier           = "STANDARD_HA"
  memory_size_gb = 1
  region         = var.region
  
  authorized_network = google_compute_network.vpc.id
  
  redis_version = "REDIS_7_0"
}

# Cloud Storage (for file uploads)
resource "google_storage_bucket" "uploads" {
  name     = "agri-platform-uploads"
  location = var.region
  
  uniform_bucket_level_access = true
  
  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type = "Delete"
    }
  }
}

# Outputs
output "database_connection" {
  value     = google_sql_database_instance.postgres.connection_name
  sensitive = true
}

output "redis_host" {
  value = google_redis_instance.cache.host
}

output "gke_cluster_name" {
  value = google_container_cluster.primary.name
}
```

---

## 12. Development Workflow

### 12.1 Git Workflow

**Branching Strategy (GitFlow)**
```
main          ─────────────────────────────► (production)
               │                      ▲
               │                      │ merge
develop   ─────┴──────────────────────┴───► (staging)
           │   │              │       │
           │   ▼              ▼       │
feature/*  │   ───────────    ───────────
           │   feature-123    feature-456
           │
           ▼
hotfix/*   ───────────
           hotfix-789
```

**Commit Message Convention**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding tests
- `chore`: Maintenance tasks

**Examples:**
```bash
feat(order): add order cancellation endpoint

Implement REST API endpoint for order cancellation with
proper validation and status checks.

Closes #123

fix(product): fix stock reservation race condition

Use pessimistic locking to prevent race conditions when
multiple orders try to reserve the same product simultaneously.

Fixes #456
```

### 12.2 Code Review Checklist
```markdown
## Code Review Checklist

### Functionality
- [ ] Code does what it's supposed to do
- [ ] Edge cases are handled
- [ ] Error handling is appropriate
- [ ] No obvious bugs

### Code Quality
- [ ] Code is readable and maintainable
- [ ] Follows Go best practices and idioms
- [ ] No code duplication
- [ ] Functions are small and focused
- [ ] Variable/function names are clear

### Testing
- [ ] Unit tests are included
- [ ] Tests cover edge cases
- [ ] Test coverage is adequate (>80%)
- [ ] Tests are meaningful (not just for coverage)

### Performance
- [ ] No obvious performance issues
- [ ] Database queries are optimized
- [ ] Proper use of indexes
- [ ] Caching where appropriate

### Security
- [ ] Input validation is proper
- [ ] No SQL injection vulnerabilities
- [ ] Authentication/authorization checks
- [ ] Sensitive data is not logged

### Documentation
- [ ] Code is well-commented
- [ ] API documentation updated
- [ ] README updated if needed
- [ ] Breaking changes documented
```

### 12.3 Local Development Setup
```bash
# 1. Clone repository
git clone https://github.com/yourusername/agri-platform.git
cd agri-platform

# 2. Install dependencies
go mod download

# 3. Setup environment variables
cp .env.example .env
# Edit .env with your local settings

# 4. Start infrastructure (Docker Compose)
docker-compose up -d postgres redis rabbitmq

# 5. Run database migrations
make migrate-up

# 6. Seed test data
make seed

# 7. Run service
make run-api-gateway

# Or run all services
make run-all
```

**Makefile**
```makefile
.PHONY: help build test run-all migrate-up migrate-down seed clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build all services
	@echo "Building services..."
	@cd services/api-gateway && go build -o ../../bin/api-gateway ./cmd/api
	@cd services/user-service && go build -o ../../bin/user-service ./cmd/server
	@cd services/product-service && go build -o ../../bin/product-service ./cmd/server
	@cd services/order-service && go build -o ../../bin/order-service ./cmd/server

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests
	go test -v -tags=integration ./tests/integration/...

lint: ## Run linter
	golangci-lint run ./...

migrate-up: ## Run database migrations
	migrate -path db/migrations -database "postgresql://agriuser:agripass@localhost:5432/agri_platform?sslmode=disable" up

migrate-down: ## Rollback last migration
	migrate -path db/migrations -database "postgresql://agriuser:agripass@localhost:5432/agri_platform?sslmode=disable" down 1

seed: ## Seed database with test data
	go run scripts/seed/main.go

run-api-gateway: ## Run API Gateway
	cd services/api-gateway && go run cmd/api/main.go

run-all: ## Run all services
	docker-compose up

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

proto: ## Generate protobuf code
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/**/*.proto

docker-build: ## Build Docker images
	docker-compose build

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f
```

---

## Conclusion

This specification document provides a comprehensive technical blueprint for the Mini Agri Supply Chain Platform. It covers all aspects from system architecture to deployment strategies, ensuring alignment with the job requirements at XXX and other backend engineering positions.

**Key Highlights:**
- ✅ Microservices architecture with 6 core services
- ✅ Golang with PostgreSQL (primary requirements)
- ✅ RESTful, gRPC, and GraphQL APIs
- ✅ Real-time tracking with WebSocket
- ✅ Event-driven architecture with RabbitMQ
- ✅ Comprehensive testing strategy (80%+ coverage)
- ✅ Cloud-native deployment (GCP/AWS)
- ✅ Complete CI/CD pipeline
- ✅ Production-ready monitoring & observability
- ✅ Security best practices

**Implementation Timeline:**
- Phase 1 (Weeks 1-3): Core services (User, Product, Order)
- Phase 2 (Weeks 4-5): Tracking & Analytics services
- Phase 3 (Weeks 6-7): Advanced features (gRPC, GraphQL, Events)
- Phase 4 (Week 8): Deployment & monitoring
- Phase 5 (Week 9): Documentation & polish

This project demonstrates senior-level backend engineering skills and directly maps to the agritech domain, making it an ideal portfolio piece for XXX applications.