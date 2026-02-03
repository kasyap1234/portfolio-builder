package repository

import (
	"context"
	"errors"
	"time"

	"smart-alert/internal/domain/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AllocatorRepository struct {
	DB *pgxpool.Pool
}

func NewAllocatorRepository(db *pgxpool.Pool) *AllocatorRepository {
	return &AllocatorRepository{DB: db}
}

// AddHolding adds or updates a portfolio holding
func (r *AllocatorRepository) AddHolding(ctx context.Context, assetID int64, quantity, avgPrice, totalInvested float64, category string) error {
	query := `
		INSERT INTO portfolio_holdings (asset_id, quantity, average_buy_price, total_invested, asset_category)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (asset_id) DO UPDATE SET
			quantity = portfolio_holdings.quantity + EXCLUDED.quantity,
			average_buy_price = (portfolio_holdings.total_invested + EXCLUDED.total_invested) / 
				NULLIF(portfolio_holdings.quantity + EXCLUDED.quantity, 0),
			total_invested = portfolio_holdings.total_invested + EXCLUDED.total_invested,
			updated_at = CURRENT_TIMESTAMP`

	_, err := r.DB.Exec(ctx, query, assetID, quantity, avgPrice, totalInvested, category)
	return err
}

// GetHoldingsByCategory retrieves holdings by category
func (r *AllocatorRepository) GetHoldingsByCategory(ctx context.Context, category string) (float64, error) {
	query := `SELECT COALESCE(SUM(total_invested), 0) FROM portfolio_holdings WHERE asset_category = $1`
	var total float64
	err := r.DB.QueryRow(ctx, query, category).Scan(&total)
	return total, err
}

// GetDebtReserve fetches the current debt reserve
func (r *AllocatorRepository) GetDebtReserve(ctx context.Context) (*models.DebtReserve, error) {
	query := `
		SELECT id, total_amount, deployed_amount, available_amount, last_deployment_date, updated_at
		FROM debt_reserve
		LIMIT 1`

	var d models.DebtReserve
	err := r.DB.QueryRow(ctx, query).Scan(&d.ID, &d.TotalAmount, &d.DeployedAmount, &d.AvailableAmount, &d.LastDeploymentDate, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// UpdateDebtReserve updates the total debt reserve amount
func (r *AllocatorRepository) UpdateDebtReserve(ctx context.Context, totalAmount float64) error {
	query := `UPDATE debt_reserve SET total_amount = $1, updated_at = CURRENT_TIMESTAMP WHERE id = 1`
	_, err := r.DB.Exec(ctx, query, totalAmount)
	return err
}

// AddToDebtReserve adds to the total debt reserve
func (r *AllocatorRepository) AddToDebtReserve(ctx context.Context, amount float64) error {
	query := `UPDATE debt_reserve SET total_amount = total_amount + $1, updated_at = CURRENT_TIMESTAMP WHERE id = 1`
	_, err := r.DB.Exec(ctx, query, amount)
	return err
}

// DeployFromDebtReserve marks an amount as deployed from debt reserve
func (r *AllocatorRepository) DeployFromDebtReserve(ctx context.Context, amount float64) error {
	query := `
		UPDATE debt_reserve 
		SET deployed_amount = deployed_amount + $1, 
			last_deployment_date = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1 AND (total_amount - deployed_amount) >= $1`

	result, err := r.DB.Exec(ctx, query, amount, time.Now())
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInsufficientDebtReserve
	}
	return nil
}

// GetAssetIDBySymbol retrieves asset ID by symbol
func (r *AllocatorRepository) GetAssetIDBySymbol(ctx context.Context, symbol string) (int64, error) {
	query := `SELECT id FROM assets WHERE symbol = $1`
	var id int64
	err := r.DB.QueryRow(ctx, query, symbol).Scan(&id)
	return id, err
}

// ErrInsufficientDebtReserve is returned when trying to deploy more than available
var ErrInsufficientDebtReserve = errors.New("insufficient debt reserve available")
