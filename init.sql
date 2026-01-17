-- Supafirehose Demo Database Schema
-- Run: psql -h localhost -U postgres -d pooler_demo -f init.sql
-- ============================================
-- Simple Scenario: Basic users table
-- ============================================
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ANALYZE users;
-- ============================================
-- JSONB Scenario: Table with JSONB payload
-- ============================================
CREATE TABLE IF NOT EXISTS jsonb_data (
    id BIGSERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Create GIN index on JSONB column for efficient queries
CREATE INDEX IF NOT EXISTS idx_jsonb_data_payload ON jsonb_data USING GIN (payload);
ANALYZE jsonb_data;
-- ============================================
-- Wide Scenario: Table with many columns
-- ============================================
CREATE TABLE IF NOT EXISTS wide_data (
    id BIGSERIAL PRIMARY KEY,
    col_01 VARCHAR(255),
    col_02 VARCHAR(255),
    col_03 VARCHAR(255),
    col_04 VARCHAR(255),
    col_05 VARCHAR(255),
    col_06 VARCHAR(255),
    col_07 VARCHAR(255),
    col_08 VARCHAR(255),
    col_09 VARCHAR(255),
    col_10 VARCHAR(255),
    col_11 VARCHAR(255),
    col_12 VARCHAR(255),
    col_13 VARCHAR(255),
    col_14 VARCHAR(255),
    col_15 VARCHAR(255),
    col_16 VARCHAR(255),
    col_17 VARCHAR(255),
    col_18 VARCHAR(255),
    col_19 VARCHAR(255),
    col_20 VARCHAR(255),
    int_01 INTEGER,
    int_02 INTEGER,
    int_03 INTEGER,
    int_04 INTEGER,
    int_05 INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Seed with 100,000 rows of wide data
INSERT INTO wide_data (
        col_01,
        col_02,
        col_03,
        col_04,
        col_05,
        col_06,
        col_07,
        col_08,
        col_09,
        col_10,
        col_11,
        col_12,
        col_13,
        col_14,
        col_15,
        col_16,
        col_17,
        col_18,
        col_19,
        col_20,
        int_01,
        int_02,
        int_03,
        int_04,
        int_05
    )
SELECT 'col01_' || i,
    'col02_' || i,
    'col03_' || i,
    'col04_' || i,
    'col05_' || i,
    'col06_' || i,
    'col07_' || i,
    'col08_' || i,
    'col09_' || i,
    'col10_' || i,
    'col11_' || i,
    'col12_' || i,
    'col13_' || i,
    'col14_' || i,
    'col15_' || i,
    'col16_' || i,
    'col17_' || i,
    'col18_' || i,
    'col19_' || i,
    'col20_' || i,
    i % 1000,
    i % 500,
    i % 250,
    i % 100,
    i % 50
FROM generate_series(1, 100000) AS i ON CONFLICT DO NOTHING;
ANALYZE wide_data;
-- ============================================
-- FK Scenario: Tables with foreign key lookup
-- ============================================
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS items CASCADE;
CREATE TABLE IF NOT EXISTS categories (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id BIGINT NOT NULL REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ANALYZE categories;
ANALYZE items;