package repository

import (
	"context"
	"time"

	"smart-alert/internal/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SnapshotRepository struct {
	DB *pgxpool.Pool
}

func NewSnapshotRepository(db *pgxpool.Pool) *SnapshotRepository {
	return &SnapshotRepository{DB: db}
}

// SaveSnapshot saves an allocation snapshot
func (r *SnapshotRepository) SaveSnapshot(ctx context.Context, snapshot *models.AllocationSnapshot) error {
	query := `
		INSERT INTO allocation_snapshots (snapshot_date, total_equity_amount, total_debt_amount, equity_allocation_pct, debt_allocation_pct, market_regime, nifty_dma_distance)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (snapshot_date) DO UPDATE SET
			total_equity_amount = EXCLUDED.total_equity_amount,
			total_debt_amount = EXCLUDED.total_debt_amount,
			equity_allocation_pct = EXCLUDED.equity_allocation_pct,
			debt_allocation_pct = EXCLUDED.debt_allocation_pct,
			market_regime = EXCLUDED.market_regime,
			nifty_dma_distance = EXCLUDED.nifty_dma_distance
		RETURNING id, created_at`

	return r.DB.QueryRow(ctx, query,
		snapshot.SnapshotDate,
		snapshot.TotalEquityAmount,
		snapshot.TotalDebtAmount,
		snapshot.EquityAllocationPct,
		snapshot.DebtAllocationPct,
		snapshot.MarketRegime,
		snapshot.NiftyDMADistance,
	).Scan(&snapshot.ID, &snapshot.CreatedAt)
}

// GetSnapshotsByDateRange retrieves snapshots within a date range
func (r *SnapshotRepository) GetSnapshotsByDateRange(ctx context.Context, from, to time.Time) ([]models.AllocationSnapshot, error) {
	query := `
		SELECT id, snapshot_date, total_equity_amount, total_debt_amount, equity_allocation_pct, debt_allocation_pct, market_regime, nifty_dma_distance, created_at
		FROM allocation_snapshots
		WHERE snapshot_date >= $1 AND snapshot_date <= $2
		ORDER BY snapshot_date DESC`

	rows, err := r.DB.Query(ctx, query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.AllocationSnapshot
	for rows.Next() {
		var s models.AllocationSnapshot
		if err := rows.Scan(&s.ID, &s.SnapshotDate, &s.TotalEquityAmount, &s.TotalDebtAmount,
			&s.EquityAllocationPct, &s.DebtAllocationPct, &s.MarketRegime, &s.NiftyDMADistance, &s.CreatedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// GetLatestSnapshot returns the most recent allocation snapshot
func (r *SnapshotRepository) GetLatestSnapshot(ctx context.Context) (*models.AllocationSnapshot, error) {
	query := `
		SELECT id, snapshot_date, total_equity_amount, total_debt_amount, equity_allocation_pct, debt_allocation_pct, market_regime, nifty_dma_distance, created_at
		FROM allocation_snapshots
		ORDER BY snapshot_date DESC
		LIMIT 1`

	var s models.AllocationSnapshot
	err := r.DB.QueryRow(ctx, query).Scan(&s.ID, &s.SnapshotDate, &s.TotalEquityAmount, &s.TotalDebtAmount,
		&s.EquityAllocationPct, &s.DebtAllocationPct, &s.MarketRegime, &s.NiftyDMADistance, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
