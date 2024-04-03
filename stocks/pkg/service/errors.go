package service

import "errors"

// StockNotFoundError.
var ErrStockNotFound = errors.New("stock not found")

// StockNotFoundError.
var ErrInvalidUpdateRequest = errors.New("invalid update request")
