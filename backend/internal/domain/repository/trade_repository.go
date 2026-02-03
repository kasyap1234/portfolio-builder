package repository

import (
	"context"
	"time"

	"smart-alert/internal/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TradeRepository struct {
	DB *pgxpool.Pool
}

func NewTradeRepository(db *pgxpool.Pool) *TradeRepository {
	return &TradeRepository{DB: db}
}

// RecordTrade inserts a new trade log
func (r *TradeRepository) RecordTrade(ctx context.Context, trade *models.TradeLog) error {
	query := `
		INSERT INTO trade_logs (asset_id, transaction_date, transaction_type, amount, units, price_per_unit, asset_category, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	return r.DB.QueryRow(ctx, query,
		trade.AssetID,
		trade.TransactionDate,
		trade.TransactionType,
		trade.Amount,
		trade.Units,
		trade.PricePerUnit,
		trade.AssetCategory,
		trade.Notes,
	).Scan(&trade.ID, &trade.CreatedAt)
}

// GetTradesByDateRange retrieves trade logs within a date range
func (r *TradeRepository) GetTradesByDateRange(ctx context.Context, from, to time.Time) ([]models.TradeLog, error) {
	query := `
		SELECT id, asset_id, transaction_date, transaction_type, amount, units, price_per_unit, asset_category, notes, created_at
		FROM trade_logs
		WHERE transaction_date >= $1 AND transaction_date <= $2
		ORDER BY transaction_date DESC`

	rows, err := r.DB.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.TradeLog
	for rows.Next() {
		var t models.TradeLog
		if err := rows.Scan(&t.ID, &t.AssetID, &t.TransactionDate, &t.TransactionType,
			&t.Amount, &t.Units, &t.PricePerUnit, &t.AssetCategory, &t.Notes, &t.CreatedAt); err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

// GetTradesByAsset retrieves all trades for a specific asset
func (r *TradeRepository) GetTradesByAsset(ctx context.Context, assetID int64) ([]models.TradeLog, error) {
	query := `
		SELECT id, asset_id, transaction_date, transaction_type, amount, units, price_per_unit, asset_category, notes, created_at
		FROM trade_logs
		WHERE asset_id = $1
		ORDER BY transaction_date DESC`

	rows, err := r.DB.Query(ctx, query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.TradeLog
	for rows.Next() {
		var t models.TradeLog
		if err := rows.Scan(&t.ID, &t.AssetID, &t.TransactionDate, &t.TransactionType,
			&t.Amount, &t.Units, &t.PricePerUnit, &t.AssetCategory, &t.Notes, &t.CreatedAt); err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

// GetTradesByCategory retrieves all trades for a specific category (EQUITY, DEBT, ETF, MF)
func (r *TradeRepository) GetTradesByCategory(ctx context.Context, category string, from, to time.Time) ([]models.TradeLog, error) {
	query := `
		SELECT id, asset_id, transaction_date, transaction_type, amount, units, price_per_unit, asset_category, notes, created_at
		FROM trade_logs
		WHERE asset_category = $1 AND transaction_date >= $2 AND transaction_date <= $3
		ORDER BY transaction_date DESC`

	rows, err := r.DB.Query(ctx, query, category, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []models.TradeLog
	for rows.Next() {
		var t models.TradeLog
		if err := rows.Scan(&t.ID, &t.AssetID, &t.TransactionDate, &t.TransactionType,
			&t.Amount, &t.Units, &t.PricePerUnit, &t.AssetCategory, &t.Notes, &t.CreatedAt); err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

// GetTotalInvestedByCategory returns total amount invested in a category
func (r *TradeRepository) GetTotalInvestedByCategory(ctx context.Context, category string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(CASE WHEN transaction_type = 'BUY' THEN amount ELSE -amount END), 0)
		FROM trade_logs
		WHERE asset_category = $1`

	var total float64
	err := r.DB.QueryRow(ctx, query, category).Scan(&total)
	return total, err
}
