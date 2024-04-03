CREATE TABLE IF NOT EXISTS user_stocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    stockId INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    FOREIGN KEY (stockId) REFERENCES stocks(id),
    UNIQUE (username, stockId)
);
