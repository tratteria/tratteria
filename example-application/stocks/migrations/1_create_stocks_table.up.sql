CREATE TABLE IF NOT EXISTS stocks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    exchange TEXT NOT NULL,
    currentPrice REAL NOT NULL,
    totalAvailableShares INTEGER NOT NULL
);