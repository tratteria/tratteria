export interface Stock {
    id: string;
    symbol: string;
    name: string;
    exchange: string;
    currentPrice: number;
    totalAvailableShares: number;
    holdings: number;
  }