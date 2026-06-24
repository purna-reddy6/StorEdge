# StorEdge — Universal Multi-Tenant Warehousing Marketplace

> India's first unified platform connecting agricultural cold storages, commercial warehouses, and e-commerce micro-fulfillment hubs with farmers, traders, and logistics providers.

## What StorEdge Does

- **Marketplace:** Book warehouse space by the pallet, hour, or month — cold storage to Grade-A commercial hubs
- **IoT Telemetry:** Real-time temperature, humidity, ethylene monitoring across all facilities
- **e-NWR FinTech:** Convert stored crops into digital collateral; get post-harvest loans from partner banks instantly
- **AI Pricing:** Dynamic pallet-rate pricing based on real-time occupancy, seasonal demand, and commodity spot prices
- **Logistics Dispatch:** Match shipments with local freelance fleets and enterprise carriers

## Quick Start (Local Development)

### Prerequisites
- Docker + Docker Compose
- Node.js v18+
- Go 1.22+
- Java 21+
- Python 3.11+

### Run All Services

```bash
# 1. Copy environment config
cp .env.example .env
# Fill in your API keys (Mapbox, etc.)

# 2. Start infrastructure (postgres, redis, kafka, elasticsearch)
docker-compose up -d postgres redis kafka elasticsearch

# 3. Run database migrations
./scripts/migrate.sh

# 4. Start all microservices
docker-compose up

# 5. Open web portal
open http://localhost:3000
```

### Environment Variables Required

| Variable | Description |
|---|---|
| `MAPBOX_TOKEN` | Mapbox public token for interactive map |
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `JWT_SECRET` | JWT signing secret (min 32 chars) |
| `KAFKA_BROKERS` | Kafka broker addresses |
| `ELASTICSEARCH_URL` | Elasticsearch endpoint |

## Monorepo Structure

```
storedge/
├── apps/
│   ├── web/          # React + Tailwind — Trader & Operator portal
│   └── mobile/       # React Native — Farmer Android app
├── services/
│   ├── search-match/ # Go + gRPC — Core matching engine
│   ├── inventory-wms/# Java Spring Boot — WMS & pallet ledger
│   ├── financing-enwrs/ # Go — e-NWR & bank gateway
│   ├── iot-gateway/  # Go — Sensor MQTT ingestion
│   └── ai-engine/    # Python FastAPI — ML pricing & recommender
├── packages/proto/   # Shared Protobuf definitions
├── infrastructure/   # K8s, Terraform, Docker
└── PROJECT_LOG.md    # Live build progress log
```

## Architecture

```
[Farmer App] [Trader Web] [Operator Portal]
       ↓           ↓              ↓
   [Cloudflare WAF / API Gateway :8080]
       ↓           ↓              ↓
[search-match] [inventory-wms] [financing-enwrs]
   (Go/gRPC)   (Java/REST)     (Go/OAuth2)
       ↓           ↓              ↓
[PostgreSQL+PostGIS] [Redis] [Elasticsearch]
              ↓
       [Apache Kafka]
              ↓
    [IoT Gateway] ← [Sensor Network]
              ↓
       [AI Engine]
```

## Go-to-Market

- **Phase 1:** Agra potato belt — 50 cold storage facilities, 1,500 farmers
- **Phase 2:** UP + Gujarat + Maharashtra — textile + e-commerce hubs
- **Phase 3:** Pan-India + cross-border pilots

## Progress

See [PROJECT_LOG.md](./PROJECT_LOG.md) for detailed build progress, decisions, and next steps.

## Company

**PN StorEdge Technologies Private Limited**  
Building the AWS for physical goods storage.
