package stockserrors

import "errors"

var ErrStockNotFound = errors.New("stock not found")

var ErrInvalidUpdateRequest = errors.New("invalid update request")
