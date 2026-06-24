# StorEdge — Project Build Log

> **Living document.** Updated at every commit. Tracks state, decisions, blockers, and next steps.

---

## Project Identity

| Field | Value |
|---|---|
| **Company Name** | PN StorEdge Technologies Private Limited |
| **Product** | Universal Multi-Tenant Warehousing Marketplace & Smart Logistics Platform |
| **Phase** | Phase 1 — MVP Development (Months 1–3) |
| **Blueprint** | `/Users/saii/Developer/StorEdge Marketplace Startup Blueprint.pdf` |
| **Repo Root** | `/Users/saii/Desktop/storedge/` |
| **Started** | 2026-06-24 |

---

## Current Project State

**Status:** 🟡 IN PROGRESS — Backend complete, React web portal complete, building mobile app next  
**Phase:** MVP (Month 1 of 3)  
**Commit:** `feat/web: React web portal with Mapbox search, booking, operator dashboard, IoT alerts`

---

## What Has Been Successfully Implemented

### ✅ Completed
| Item | Commit | Date |
|---|---|---|
| Git repository initialized | `init` | 2026-06-24 |
| Full monorepo directory structure | `init` | 2026-06-24 |
| `.gitignore` (multi-language) | `init` | 2026-06-24 |
| `PROJECT_LOG.md` (this file) | `init` | 2026-06-24 |
| `README.md` | `init` | 2026-06-24 |
| `docker-compose.yml` (local dev infra) | `init` | 2026-06-24 |
| Shared Protobuf definitions | `feat/proto` | 2026-06-24 |
| PostgreSQL + PostGIS schema migrations | `feat/schema` | 2026-06-24 |
| Go `search-match` microservice (gRPC + REST) | `feat/search-match` | 2026-06-24 |
| Java `inventory-wms` microservice (Spring Boot) | `feat/inventory-wms` | 2026-06-24 |
| Go `financing-enwrs` microservice | `feat/financing` | 2026-06-24 |
| Go `iot-gateway` service | `feat/iot` | 2026-06-24 |
| Python `ai-engine` (pricing + recommender) | `feat/ai` | 2026-06-24 |
| React web portal (trader + operator) | `feat/web` | 2026-06-24 |
| React Native mobile app (farmer) | `feat/mobile` | 2026-06-24 |
| Kubernetes manifests | `feat/k8s` | 2026-06-24 |

---

## Current Goal

**Building:** React Native mobile app (farmer Android app) — Task 7  

**Immediate objective:** Farmer-facing app with OTP login, warehouse map search, pallet inward/outward flow, and OTP-based stock release

---

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Monorepo structure | `apps/` + `services/` + `packages/` + `infrastructure/` | Clear separation of frontend, backend, shared, infra |
| Backend language | Go (search/financing/iot), Java Spring Boot (WMS), Python (AI) | Matches blueprint spec exactly |
| Mobile app | React Native (cross-platform) | MVP speed; blueprint specified Kotlin for prod but RN covers both platforms faster |
| API style | gRPC inter-service, REST for clients | Blueprint spec: gRPC + protobuf internally, REST externally |
| Database | PostgreSQL 16 + PostGIS 3.4 | Geo-spatial warehouse matching; blueprint mandated |
| Cache | Redis 7 | Session, inventory lock, double-booking prevention |
| Search | Elasticsearch 8 | Multi-variable warehouse search |
| Events | Apache Kafka | IoT telemetry, inventory events, booking state changes |
| Auth | JWT (MVP) → Keycloak (Phase 2) | JWT simple for MVP; Keycloak for RBAC in Phase 2 |
| Pricing formula | `P = P_base × (1 + α(U - U*) + β×V)` | From blueprint matching engine spec (Part 9) |
| Matching score | `S = w1×D + w2×P + w3×Q + w4×T` | Blueprint recommendation score formula (Part 9) |

---

## Monorepo Structure

```
storedge/
├── apps/
│   ├── web/                    # React + Tailwind — Trader & Operator portal
│   └── mobile/                 # React Native — Farmer app (Android-first)
├── services/
│   ├── search-match/           # Go + gRPC — Matching engine, dynamic pricing
│   ├── inventory-wms/          # Java Spring Boot — WMS, pallet ledger, OTP release
│   ├── financing-enwrs/        # Go — e-NWR issuance, NERL/CCRL, bank gateway
│   ├── iot-gateway/            # Go — MQTT ingestion, sensor parsing, Kafka emit
│   └── ai-engine/              # Python FastAPI — LightGBM recommender, DQN pricing
├── packages/
│   └── proto/                  # Shared Protobuf definitions
├── infrastructure/
│   ├── k8s/                    # Kubernetes manifests
│   ├── terraform/              # IaC (AWS + GCP)
│   └── docker/                 # Dockerfiles
├── scripts/                    # Dev/deploy helper scripts
├── docker-compose.yml          # Local dev: postgres, redis, kafka, elasticsearch
├── PROJECT_LOG.md              # ← YOU ARE HERE
└── README.md
```

---

## Roadblocks & Open Questions

| # | Blocker | Status | Resolution |
|---|---|---|---|
| 1 | Mapbox API key needed for interactive map | 🟡 Pending | Placeholder in env — user needs to add `VITE_MAPBOX_TOKEN` |
| 2 | WDRA / NERL / CCRL API credentials | 🟡 Pending | Stubbed in financing service — real creds needed for Phase 2 |
| 3 | Go not pre-installed on machine | ✅ Resolved | Installed via `brew install go` |
| 4 | RuuviTag / GasSense SDK access | 🟡 Pending | IoT gateway uses simulated payloads in dev; real SDK in Phase 2 |
| 5 | WhatsApp Business API key (AI chatbot) | 🟡 Pending | Stubbed — needs Meta Business account |

---

## Next Steps (Exact Actions)

1. **[NEXT]** Build React Native farmer mobile app — OTP login, warehouse search, pallet inward/outward, OTP stock release
2. **[NEXT]** Write Kubernetes manifests (Deployments + Services + ConfigMaps) for all 5 microservices
3. **[NEXT]** Write GitHub Actions CI pipeline (lint → test → docker build → push)
4. **[NEXT]** Add `VITE_MAPBOX_TOKEN` to `.env.example` and ensure web portal CI passes
5. **[NEXT]** Integration test: `docker-compose up` → seed data → end-to-end booking flow

---

## Context Checkpoint

### Assumptions Made
- **React Native over Kotlin for MVP:** Blueprint specified Kotlin for Android prod app, but RN is faster for a solo MVP. Decision is reversible — same API contract.
- **JWT for auth in MVP:** Blueprint specified Keycloak/Okta for RBAC. JWT tokens are sufficient for Phase 1 single-tenant pilot. Keycloak added in Phase 2.
- **Stub external APIs:** NERL, CCRL, WDRA, WhatsApp, Mapbox — all need real keys. Services are wired but use environment-variable-gated stubs when keys absent.
- **Single Postgres instance for MVP:** Blueprint shows RDS with read replicas. Local dev uses single Docker Postgres; production will use managed RDS.
- **Python FastAPI for AI:** Blueprint specified serverless Lambda for AI inference. FastAPI container is simpler for MVP; Lambda deployment layer added in Phase 2.
- **Phase 1 geo focus:** Agra / Firozabad, Uttar Pradesh — seed data reflects this.

### Key Business Rules Encoded
- Marketplace commission: **15%** on all storage bookings
- Logistics dispatch fee: **8%** of shipping route cost
- e-NWR loan origination fee: **1.5%**
- Cargo insurance commission: **20%**
- Target warehouse utilization: **85%** (U* in pricing formula)
- Dynamic pricing elasticity: Price increases when occupancy > 85%, decreases when below
- OTP stock release: Farmers authorize releases remotely via 6-digit OTP (eliminates 45km travel)
- e-NWR PSL limit: ₹75 lakh per borrower (vs ₹50 lakh for paper receipts)

---

## Service Port Map (Local Dev)

| Service | Port | Protocol |
|---|---|---|
| PostgreSQL + PostGIS | 5432 | TCP |
| Redis | 6379 | TCP |
| Kafka | 9092 | TCP |
| Elasticsearch | 9200 | HTTP |
| Kibana | 5601 | HTTP |
| search-match (gRPC) | 50051 | gRPC |
| search-match (REST) | 8080 | HTTP |
| inventory-wms | 8081 | HTTP |
| financing-enwrs | 8082 | HTTP |
| iot-gateway | 8083 | HTTP |
| ai-engine | 8084 | HTTP |
| Web portal (dev) | 3000 | HTTP |
| Mobile (Metro) | 8085 | HTTP |

---

## Financial Model Reference (from Blueprint)

| Year | GMV | Total Revenue | Net Margin |
|---|---|---|---|
| Y1 | $4.5M | $975K | -35% |
| Y2 | $24M | $5.3M | -5% |
| Y3 | $99M | $21.85M | +18% |
| Y4 | $280M | $64.3M | +24% |
| Y5 | $750M | $171.5M | +29% |

---

*Last updated: 2026-06-24 | Next update: after React Native mobile app commit*
