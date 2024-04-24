export interface TransactionDetails {
    transactionID: string;
    operation: string;
    stockName: string;
    stockSymbol: string;
    stockID: number;
    stockExchange: string;
    stockPrice: number;
    quantity: number;
    totalValue: number;
  }