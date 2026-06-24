-- Migration 001: Enable required PostgreSQL extensions
-- PostGIS for geospatial queries, uuid-ossp for UUID generation

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- Trigram matching for full-text search
