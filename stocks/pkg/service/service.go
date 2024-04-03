package service

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

type StockSearchItem struct {
	Id     int    `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type StockItem struct {
	Id                   int     `json:"id"`
	Symbol               string  `json:"symbol"`
	Name                 string  `json:"name"`
	Exchange             string  `json:"exchange"`
	CurrentPrice         float64 `json:"currentPrice"`
	TotalAvailableShares int     `json:"totalAvailableShares"`
	Holdings             int     `json:"holdings"`
}

type UpdateType string

const (
	Buy  UpdateType = "Buy"
	Sell UpdateType = "Sell"
)

type UpdateDetails struct {
	Operation     UpdateType `json:"operation"`
	StockName     string     `json:"stockName"`
	StockID       int        `json:"stockID"`
	StockExchange string     `json:"stockExchange"`
	StockSymbol   string     `json:"stockSymbol"`
	Quantity      int        `json:"quantity"`
	StockPrice    float64    `json:"stockPrice"`
}

type Holding struct {
	StockID              string  `json:"stockID"`
	StockSymbol          string  `json:"stockSymbol"`
	StockName            string  `json:"stockName"`
	StockExchange        string  `json:"stockExchange"`
	TotalAvailableShares int     `json:"totalAvailableShares"`
	Quantity             int     `json:"quantity"`
	CurrentPrice         float64 `json:"currentPrice"`
	TotalValue           float64 `json:"totalValue"`
}

type Holdings struct {
	TotalHoldings int       `json:"totalHoldings"`
	TotalValue    float64   `json:"totalValue"`
	Holdings      []Holding `json:"holdings"`
}

func (s *Service) SearchStocks(query string, limit int) ([]StockSearchItem, error) {
	var stocks []StockSearchItem

	sqlStatement := `SELECT id, symbol, name FROM stocks WHERE name LIKE ? OR symbol LIKE ? LIMIT ?`

	rows, err := s.DB.Query(sqlStatement, "%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		s.Logger.Error("Error executing stock search query.", zap.Error(err))

		return nil, fmt.Errorf("error executing search query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var stock StockSearchItem
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

func (s *Service) GetStockDetails(username string, id int) (StockItem, error) {
	var stock StockItem

	sqlStatement := `
	SELECT s.id, s.symbol, s.name, s.exchange, s.currentPrice, s.totalAvailableShares, COALESCE(us.quantity, 0) AS holdings
	FROM stocks s
	LEFT JOIN user_stocks us ON s.id = us.stockId AND us.username = ?
	WHERE s.id = ?
	`

	err := s.DB.QueryRow(sqlStatement, username, id).Scan(&stock.Id, &stock.Symbol, &stock.Name, &stock.Exchange, &stock.CurrentPrice, &stock.TotalAvailableShares, &stock.Holdings)

	if err != nil {
		if err == sql.ErrNoRows {
			s.Logger.Error("Stock or user holdings not found.", zap.Int("stock_id", id), zap.String("username", username))

			return StockItem{}, fmt.Errorf("stock or user holdings not found for stock ID %d and username %s: %w", id, username, err)
		}

		s.Logger.Error("Error fetching stock details.", zap.Int("stock_id", id), zap.String("username", username), zap.Error(err))

		return StockItem{}, fmt.Errorf("error fetching stock details by ID for username %s: %v", username, err)
	}

	return stock, nil
}

func (s *Service) UpdateUserStock(username string, stock StockItem, updateType UpdateType, quantity int) (UpdateDetails, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		s.Logger.Error("Failed to start transaction.", zap.Error(err))
		return UpdateDetails{}, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer tx.Rollback()

	var existingQuantity int

	err = tx.QueryRow("SELECT quantity FROM user_stocks WHERE username = ? AND stockId = ?", username, stock.Id).Scan(&existingQuantity)
	if err != nil && err != sql.ErrNoRows {
		s.Logger.Error("Error querying existing user stock.", zap.Error(err))
		return UpdateDetails{}, fmt.Errorf("error querying existing user stock: %w", err)
	}

	newQuantity := existingQuantity
	switch updateType {
	case Buy:
		newQuantity += quantity
	case Sell:
		newQuantity -= quantity
		if newQuantity < 0 {
			return UpdateDetails{}, ErrInvalidUpdateRequest
		}
	}

	if err == sql.ErrNoRows && updateType == Buy {
		_, err = tx.Exec("INSERT INTO user_stocks (username, stockId, quantity) VALUES (?, ?, ?)", username, stock.Id, quantity)
	} else if err == nil {
		if newQuantity == 0 {
			_, err = tx.Exec("DELETE FROM user_stocks WHERE username = ? AND stockId = ?", username, stock.Id)
		} else {
			_, err = tx.Exec("UPDATE user_stocks SET quantity = ? WHERE username = ? AND stockId = ?", newQuantity, username, stock.Id)
		}
	}

	if err != nil {
		s.Logger.Error("Error updating user stocks.", zap.Error(err))
		return UpdateDetails{}, fmt.Errorf("error updating user stocks: %w", err)
	}

	var updatedAvailable int
	if updateType == Buy {
		updatedAvailable = stock.TotalAvailableShares - quantity
	} else {
		updatedAvailable = stock.TotalAvailableShares + quantity
	}

	updateStockSQL := "UPDATE stocks SET totalAvailableShares = ? WHERE id = ?"

	_, err = tx.Exec(updateStockSQL, updatedAvailable, stock.Id)
	if err != nil {
		s.Logger.Error("Error updating total available shares.", zap.Error(err))
		return UpdateDetails{}, fmt.Errorf("error updating total available shares: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.Logger.Error("Failed to commit transaction.", zap.Error(err))
		return UpdateDetails{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return UpdateDetails{
		Operation:     updateType,
		StockName:     stock.Name,
		StockID:       stock.Id,
		StockSymbol:   stock.Symbol,
		StockExchange: stock.Exchange,
		Quantity:      quantity,
		StockPrice:    stock.CurrentPrice,
	}, nil
}

func (s *Service) GetUserHoldings(username string) (Holdings, error) {
	var holdings Holdings

	sqlStatement := `
	SELECT s.id, s.symbol, s.name, s.exchange, s.totalAvailableShares, us.quantity, s.currentPrice
	FROM user_stocks us
	INNER JOIN stocks s ON us.stockId = s.id
	WHERE us.username = ?
	`

	rows, err := s.DB.Query(sqlStatement, username)
	if err != nil {
		s.Logger.Error("Error executing query to fetch user holdings.", zap.Error(err))

		return Holdings{}, fmt.Errorf("error executing query to fetch user holdings: %w", err)
	}

	defer rows.Close()

	var totalValue float64
	for rows.Next() {
		var holding Holding

		if err := rows.Scan(&holding.StockID, &holding.StockSymbol, &holding.StockName, &holding.StockExchange, &holding.TotalAvailableShares, &holding.Quantity, &holding.CurrentPrice); err != nil {
			s.Logger.Error("Error scaning holdings rows.", zap.Error(err))

			return Holdings{}, fmt.Errorf("error scaning holdings rows: %w", err)
		}

		holding.TotalValue = float64(holding.Quantity) * holding.CurrentPrice
		totalValue += holding.TotalValue
		holdings.Holdings = append(holdings.Holdings, holding)
	}

	holdings.TotalHoldings = len(holdings.Holdings)
	holdings.TotalValue = totalValue

	return holdings, nil
}
