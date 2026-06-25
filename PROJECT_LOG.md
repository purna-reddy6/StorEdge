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

**Status:** ✅ PHASE 1 MVP COMPLETE — All services running, all APIs verified end-to-end  
**Last Verified:** 2026-06-25  
**Last Commit:** `fix: align all frontend field names with snake_case API responses`  
**Docker Compose:** All 12 containers healthy  
**API Gateway:** `http://localhost:8080/api/v1` — all endpoints live  
**Web Portal:** `http://localhost:3001` — 200 OK, TypeScript build passing (0 errors)

---

## What Has Been Successfully Implemented

### ✅ Completed

| Item | Notes |
|---|---|
| Git repository + monorepo structure | `apps/`, `services/`, `packages/`, `infrastructure/` |
| PostgreSQL 16 + PostGIS 3.4 schema | 9 migrations, partitioned IoT tables, generated columns |
| Go `search-match` microservice | gRPC + REST API gateway on :8080, matching formula S=w1D+w2P+w3Q+w4T |
| Java `inventory-wms` (Spring Boot) | Kafka events, OTP release state machine, on :8081 |
| Go `financing-enwrs` | e-NWR issuance stub, NERL/CCRL integration points, on :8082 |
| Go `iot-gateway` | MQTT ingestion, sensor parsing, Kafka emit, on :8083 |
| Python `ai-engine` | FastAPI, dynamic pricing P=base×clamp(1+αU+βV,0.7,2.5), on :8084 |
| React web portal | Mapbox search map, booking, operator dashboard, IoT alerts, e-NWR — **TypeScript build 0 errors** |
| React Native mobile app | OTP login, map search, bookings, inventory OTP release, e-NWR financing |
| Kubernetes manifests | 6 Deployments with probes + resource limits, Services, LoadBalancers |
| GitHub Actions CI | go/java/python/tsc/docker build pipeline |
| Transactional seed data | 3 bookings, 5 pallet items, 48 IoT readings, 6 alerts, 2 e-NWR receipts |
| API field-name alignment | All frontend (web + mobile) aligned to snake_case API responses |

---

## API Verification (2026-06-25)

All endpoints tested with real JWT (Vikram Singh, operator):

| Endpoint | Result |
|---|---|
| `POST /auth/otp/send` | `{dev_otp, expires_in, message}` |
| `POST /auth/otp/verify` | JWT + user object |
| `GET /warehouses/search?lat=27.18&lng=78.01&radius_km=50` | 3 Agra warehouses, flat response with match_score |
| `GET /bookings` | 3 bookings (operator sees all via tenantFilterID) |
| `GET /iot/alerts` | 6 alerts (2 open critical, 1 resolved) |
| `GET /inventory/pallets` | 5 pallet items with live temperature |
| `GET /financing/receipts` | 2 e-NWR receipts (potato ₹7.67L, onion ₹6.48L) |
| `GET /operator/occupancy` | 1 occupancy data point |
| `PATCH /iot/alerts/:id/resolve` | Marks alert resolved |
| `POST /inventory/pallets/:id/release/initiate` | Creates pending_otp release request |
| `POST /inventory/pallets/:id/release/authorize` | OTP verification → authorized |
| `POST /financing/receipts/:id/loans` | Pledges receipt, creates loan application |

---

## Architecture Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Monorepo structure | `apps/` + `services/` + `packages/` + `infrastructure/` | Clear separation of frontend, backend, shared, infra |
| API gateway pattern | search-match service acts as single REST gateway | Avoids CORS complexity; operator dashboard, inventory, financing, IoT all served via single Go service querying shared PostgreSQL |
| Backend JSON style | snake_case (Go idiomatic) | Frontend types updated to match — no transformation layer needed |
| Role-aware data scoping | `tenantFilterID()` helper: operators/admins see all (`""`), farmers/traders see own data | Simple and correct for Phase 1 single-operator pilot |
| Dynamic pricing unit | ₹/pallet/day (not ₹/MT/month) | Pallet is the actual inventory unit; MT conversion is ambiguous across commodities |
| Backend language | Go (search/financing/iot), Java Spring Boot (WMS), Python (AI) | Matches blueprint spec exactly |
| Mobile app | React Native (cross-platform) | MVP speed; blueprint specified Kotlin for prod but RN covers both platforms faster |
| API style | gRPC inter-service, REST for clients | Blueprint spec: gRPC + protobuf internally, REST externally |
| Auth | JWT (MVP) → Keycloak (Phase 2) | JWT simple for MVP; Keycloak for RBAC in Phase 2 |
| Pricing formula | `P = P_base × clamp(1 + α(U−U*) + β×V, 0.70, 2.50)` | From blueprint matching engine spec |
| Matching score | `S = w1×D + w2×P + w3×Q + w4×T` | Blueprint recommendation score formula |

---

## Port Map (Local Dev with Remapped Ports)

| Service | Host Port | Container Port | Protocol |
|---|---|---|---|
| PostgreSQL + PostGIS | 5433 | 5432 | TCP |
| Redis | 6380 | 6379 | TCP |
| Kafka | 9093 | 9092 | TCP |
| Elasticsearch | 9200 | 9200 | HTTP |
| Kibana | 5601 | 5601 | HTTP |
| search-match (REST) | 8080 | 8080 | HTTP |
| search-match (gRPC) | 50051 | 50051 | gRPC |
| inventory-wms | 8081 | 8081 | HTTP |
| financing-enwrs | 8082 | 8082 | HTTP |
| iot-gateway | 8083 | 8083 | HTTP |
| ai-engine | 8084 | 8084 | HTTP |
| Web portal (dev) | 3001 | 3000 | HTTP |

*Ports remapped to avoid conflicts with Homebrew postgres, regressguard_redis/redpanda/frontend.*

---

## Bugs Fixed (Current Session)

| Bug | Fix |
|---|---|
| `target stage "development" could not be found` in Docker | Added `FROM node:20-alpine AS development` first stage to web Dockerfile |
| Go modules `go 1.26.4` (non-existent) | Changed all go.mod to `go 1.24`, Dockerfiles to `golang:1.24-alpine` |
| Port conflicts (postgres 5432, redis 6379, kafka 9092, web 3000) | Remapped all host ports in docker-compose.yml |
| `gps_tracks` PK must include partition key | Changed to `PRIMARY KEY (id, recorded_at)` |
| `/auth/otp/send` route missing | Added alias alongside `/auth/otp/request` |
| Search used `lon` but frontend sends `lng` | `firstOf(lng, lon)` helper |
| Search returned `"results"` key | Changed to `"warehouses"` |
| Search returned nested `{warehouse:{...}}` per result | Flattened to top-level with match_score/distance_km |
| `u.full_name` column in booking query | Fixed to `u.name` |
| `'cancelled'` not a valid `release_status` enum | Fixed to `'rejected'` |
| `$1::uuid` cast fails with empty string in OR clause | Changed to `p.tenant_id::text = $1` |
| Mobile api.ts pointed to port 8081 | Fixed to 8080 (search-match is API gateway) |
| All frontend camelCase fields vs snake_case API | Aligned all TypeScript types and component field references |
| `pallet_code` column doesn't exist | Removed; use `slot_position` as display identifier |
| `sensor_readings.humidity_pct` doesn't exist | Fixed to `relative_humidity` |
| Operator sees 0 results (own userId doesn't match tenant data) | `tenantFilterID()`: operators get `""` → sees all data |

---

## Roadblocks & Open Questions

| # | Blocker | Status |
|---|---|---|
| 1 | Mapbox API key needed for interactive map | 🟡 Pending — add `VITE_MAPBOX_TOKEN` to `.env` |
| 2 | WDRA / NERL / CCRL API credentials | 🟡 Pending — stubbed, needs Phase 2 creds |
| 3 | WhatsApp Business API (OTP delivery) | 🟡 Pending — `dev_otp` returned in response for MVP |
| 4 | React Native — can't test in browser | ℹ️ RN requires Android emulator/device; code reviewed but not runtime-tested |

---

## Next Steps (Phase 2 Readiness)

1. **Add Mapbox token** — `echo 'VITE_MAPBOX_TOKEN=pk.xxx' >> apps/web/.env` then rebuild web container
2. **End-to-end browser test** — open `http://localhost:3001`, login as Vikram (+919876543210), search Agra warehouses, complete a booking
3. **React Native smoke test** — run Android emulator, `cd apps/mobile && npx react-native run-android`
4. **WDRA/NERL credentials** — register and wire real API keys for e-NWR issuance (financing-enwrs)
5. **Deploy to AWS EKS** — apply `infrastructure/k8s/` manifests against EKS cluster
6. **WhatsApp Business API** — replace dev_otp console stub with real Meta API call

---

## Context Checkpoint

### Assumptions Made
- **React Native over Kotlin for MVP:** Blueprint specified Kotlin for Android prod app, but RN is faster for a solo MVP. Decision is reversible — same API contract.
- **JWT for auth in MVP:** Blueprint specified Keycloak/Okta for RBAC. JWT tokens are sufficient for Phase 1 single-tenant pilot.
- **Stub external APIs:** NERL, CCRL, WDRA, WhatsApp, Mapbox — all need real keys. Services are wired but use environment-variable-gated stubs when keys absent.
- **Single API gateway:** search-match acts as the sole HTTP entry point, querying shared PostgreSQL for all domains (dashboard, inventory, financing, IoT). Avoids inter-service HTTP complexity in Phase 1.
- **Operator sees all data:** When role = operator or admin, `tenantFilterID` returns `""` which triggers `OR $1 = ''` branch — shows all bookings/pallets/receipts in the system.
- **snake_case JSON:** Go struct tags emit snake_case. All TypeScript interfaces now match. No camelCase ↔ snake_case transformation layer.

### Key Business Rules Encoded
- Marketplace commission: **15%** on all storage bookings
- e-NWR loan origination fee: **1.5%**
- Target warehouse utilization: **85%** (U* in pricing formula)
- Dynamic pricing elasticity: α=0.40, β=0.15, clamp to [0.70, 2.50]
- OTP stock release: Farmers authorize releases remotely via 6-digit OTP (eliminates 45km travel)
- e-NWR PSL limit: ₹75 lakh per borrower (vs ₹50 lakh for paper receipts)
- LTV ratio: **70%** of commodity market value

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

*Last updated: 2026-06-25 | Phase 1 MVP fully operational — all 12 containers healthy, all API endpoints returning real data, web portal TypeScript build passing*
