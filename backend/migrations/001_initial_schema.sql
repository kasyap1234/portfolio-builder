-- +goose Up
-- +goose StatementBegin
CREATE TABLE assets (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('Index', 'MF', 'Stock', 'Cash')),
    scheme_code VARCHAR(20),
    target_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE asset_prices (
    id SERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    price DECIMAL(15, 4) NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ath_price DECIMAL(15, 4),
    pe_ratio DECIMAL(8, 2)
);

CREATE TABLE portfolio_holdings (
    id SERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    quantity DECIMAL(15, 4) NOT NULL DEFAULT 0,
    average_buy_price DECIMAL(15, 4) NOT NULL DEFAULT 0,
    total_invested DECIMAL(12, 2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

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

CREATE INDEX idx_asset_prices_asset_id ON asset_prices(asset_id);
CREATE INDEX idx_asset_prices_recorded_at ON asset_prices(recorded_at);
CREATE INDEX idx_portfolio_holdings_asset_id ON portfolio_holdings(asset_id);
CREATE INDEX idx_alerts_asset_id ON alerts(asset_id);
CREATE INDEX idx_alerts_sent_at ON alerts(sent_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_alerts_sent_at;
DROP INDEX IF EXISTS idx_alerts_asset_id;
DROP INDEX IF EXISTS idx_portfolio_holdings_asset_id;
DROP INDEX IF EXISTS idx_asset_prices_recorded_at;
DROP INDEX IF EXISTS idx_asset_prices_asset_id;

DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS portfolio_holdings;
DROP TABLE IF EXISTS asset_prices;
DROP TABLE IF EXISTS assets;
-- +goose StatementEnd
