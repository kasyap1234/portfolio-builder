-- +goose Up
-- +goose StatementBegin
INSERT INTO assets (symbol, name, type, scheme_code, target_amount) VALUES
    ('NIFTY_50', 'Nifty 50 Index', 'Index', NULL, 10000.00),
    ('PPFAS', 'PPFAS Long Term Value Fund', 'MF', '122639', 10000.00),
    ('WhiteOak', 'WhiteOak Capital Mutual Fund', 'MF', '148858', 5000.00),
    ('Helios', 'Helios Mutual Fund', 'MF', '150774', 5000.00),
    ('Debt', 'Cash/Debt Allocation', 'Cash', NULL, 5000.00);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM assets WHERE symbol IN ('NIFTY_50', 'PPFAS', 'WhiteOak', 'Helios', 'Debt');
-- +goose StatementEnd
