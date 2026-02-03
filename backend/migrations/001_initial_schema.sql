-- +goose Up
-- +goose StatementBegin

-- Core assets table
CREATE TABLE assets (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('Index', 'MF', 'Stock', 'Cash', 'ETF')),
    scheme_code VARCHAR(20),
    target_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Historical asset prices
CREATE TABLE asset_prices (
    id SERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    price DECIMAL(15, 4) NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ath_price DECIMAL(15, 4),
    pe_ratio DECIMAL(8, 2)
);

-- Current portfolio holdings
CREATE TABLE portfolio_holdings (
    id SERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL UNIQUE REFERENCES assets(id) ON DELETE CASCADE,
    quantity DECIMAL(15, 4) NOT NULL DEFAULT 0,
    average_buy_price DECIMAL(15, 4) NOT NULL DEFAULT 0,
    total_invested DECIMAL(12, 2) NOT NULL DEFAULT 0,
    asset_category VARCHAR(20) NOT NULL DEFAULT 'EQUITY' CHECK (asset_category IN ('EQUITY', 'DEBT', 'ETF', 'MF')),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Alerts and recommendations
CREATE TABLE alerts (
    id SERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL CHECK (alert_type IN ('Buy', 'Hold', 'Tactical_Bulk_Buy')),
    recommended_amount DECIMAL(12, 2) NOT NULL,
    drawdown_percentage DECIMAL(5, 2),
    pe_ratio DECIMAL(8, 2),
    psychology_note TEXT,
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    telegram_message_id VARCHAR(50)
);

-- Trade logs for all transactions
CREATE TABLE trade_logs (
    id SERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    transaction_date DATE NOT NULL,
    transaction_type VARCHAR(10) NOT NULL CHECK (transaction_type IN ('BUY', 'SELL', 'DIVIDEND')),
    amount DECIMAL(15, 4) NOT NULL,
    units DECIMAL(15, 6),
    price_per_unit DECIMAL(15, 4),
    asset_category VARCHAR(20) NOT NULL CHECK (asset_category IN ('EQUITY', 'DEBT', 'ETF', 'MF')),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Monthly allocation snapshots for historical tracking
CREATE TABLE allocation_snapshots (
    id SERIAL PRIMARY KEY,
    snapshot_date DATE NOT NULL,
    total_equity_amount DECIMAL(15, 4) NOT NULL,
    total_debt_amount DECIMAL(15, 4) NOT NULL,
    equity_allocation_pct DECIMAL(5, 2) NOT NULL,
    debt_allocation_pct DECIMAL(5, 2) NOT NULL,
    market_regime VARCHAR(20),
    nifty_dma_distance DECIMAL(8, 4),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(snapshot_date)
);

-- Debt reserve for panic buying deployment
CREATE TABLE debt_reserve (
    id SERIAL PRIMARY KEY,
    total_amount DECIMAL(15, 4) NOT NULL DEFAULT 0,
    deployed_amount DECIMAL(15, 4) NOT NULL DEFAULT 0,
    available_amount DECIMAL(15, 4) GENERATED ALWAYS AS (total_amount - deployed_amount) STORED,
    last_deployment_date DATE,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Performance indexes
CREATE INDEX idx_asset_prices_asset_id ON asset_prices(asset_id);
CREATE INDEX idx_asset_prices_recorded_at ON asset_prices(recorded_at);
CREATE INDEX idx_portfolio_holdings_asset_id ON portfolio_holdings(asset_id);
CREATE INDEX idx_alerts_asset_id ON alerts(asset_id);
CREATE INDEX idx_alerts_sent_at ON alerts(sent_at);
CREATE INDEX idx_trade_logs_asset_id ON trade_logs(asset_id);
CREATE INDEX idx_trade_logs_transaction_date ON trade_logs(transaction_date);
CREATE INDEX idx_trade_logs_asset_category ON trade_logs(asset_category);
CREATE INDEX idx_allocation_snapshots_date ON allocation_snapshots(snapshot_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_allocation_snapshots_date;
DROP INDEX IF EXISTS idx_trade_logs_asset_category;
DROP INDEX IF EXISTS idx_trade_logs_transaction_date;
DROP INDEX IF EXISTS idx_trade_logs_asset_id;
DROP INDEX IF EXISTS idx_alerts_sent_at;
DROP INDEX IF EXISTS idx_alerts_asset_id;
DROP INDEX IF EXISTS idx_portfolio_holdings_asset_id;
DROP INDEX IF EXISTS idx_asset_prices_recorded_at;
DROP INDEX IF EXISTS idx_asset_prices_asset_id;

DROP TABLE IF EXISTS debt_reserve;
DROP TABLE IF EXISTS allocation_snapshots;
DROP TABLE IF EXISTS trade_logs;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS portfolio_holdings;
DROP TABLE IF EXISTS asset_prices;
DROP TABLE IF EXISTS assets;
-- +goose StatementEnd
