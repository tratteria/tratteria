CREATE TABLE IF NOT EXISTS order_table (
    order_id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    stock_symbol TEXT NOT NULL,
    stock_name TEXT NOT NULL,
    stock_id INTEGER NOT NULL,
    stock_exchange TEXT NOT NULL,
    stock_price REAL NOT NULL,
    order_type TEXT NOT NULL CHECK(order_type IN ('Buy', 'Sell')),
    quantity INTEGER NOT NULL,
    total_value REAL NOT NULL
);
