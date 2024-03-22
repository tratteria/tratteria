package stockservice

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

type Service struct {
	DB     *sql.DB
	Logger *zap.Logger
}

func NewService(db *sql.DB, logger *zap.Logger) *Service {
	return &Service{
		DB:     db,
		Logger: logger,
	}
}

type SearchStockItem struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type StockItem struct {
	Id                   string `json:"id"`
	Symbol               string `json:"symbol"`
	Name                 string `json:"name"`
	Exchange             string `json:"exchange"`
	CurrentPrice         string `json:"currentPrice"`
	TotalAvailableShares string `json:"totalAvailableShares"`
}

func (s *Service) SearchStocks(query string, limit int) ([]SearchStockItem, error) {
	var stocks []SearchStockItem

	sqlStatement := `SELECT id, symbol, name FROM stocks WHERE name LIKE ? OR symbol LIKE ? LIMIT ?`

	rows, err := s.DB.Query(sqlStatement, "%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		s.Logger.Error("Error executing stock search query.", zap.Error(err))

		return nil, fmt.Errorf("error executing search query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var stock SearchStockItem
		if err := rows.Scan(&stock.Id, &stock.Symbol, &stock.Name); err != nil {
			s.Logger.Error("Error scanning stock rows.", zap.Error(err))

			continue
		}

		stocks = append(stocks, stock)
	}

	if err = rows.Err(); err != nil {
		s.Logger.Error("Error iterating over stock-search results.", zap.Error(err))

		return nil, fmt.Errorf("error iterating over results: %w", err)
	}

	return stocks, nil
}

func (s *Service) GetStockDetails(id int) (StockItem, error) {
	var stock StockItem

	sqlStatement := `SELECT id, symbol, name, exchange, currentPrice, totalAvailableShares FROM stocks WHERE id = ?`

	err := s.DB.QueryRow(sqlStatement, id).Scan(&stock.Id, &stock.Symbol, &stock.Name, &stock.Exchange, &stock.CurrentPrice, &stock.TotalAvailableShares)

	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Error("Stock not found.", zap.Int("stock_id", id))

			return StockItem{}, ErrStockNotFound
		}

		s.Logger.Error("Error fetching stock by id.", zap.Int("stock_id", id), zap.Error(err))

		return StockItem{}, fmt.Errorf("error fetching stock by ID: %v", err)
	}

	return stock, nil
}
