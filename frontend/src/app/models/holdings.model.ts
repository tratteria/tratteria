export interface Holding {
    stockID: string;
    stockSymbol: string;
    stockName: string;
    stockExchange: string;
    quantity: number;
    totalAvailableShares: number;
    currentPrice: number;
    totalValue: number;
  }
  
  export interface HoldingsResponse {
    totalHoldings: number;
    totalValue: number;
    holdings: Holding[];
  }
  