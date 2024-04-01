import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-portfolio',
  templateUrl: './portfolio.component.html',
  styleUrls: ['./portfolio.component.css']
})
export class PortfolioComponent implements OnInit {
  username: string = '';
  selectedStock: any = null;
  openStates: Map<string, boolean> = new Map();


  holdings = [
    {
      "stockSymbol": "AAPL",
      "stockName": "Apple Inc.",
      "stockExchange": "NASDAQ",
      "stockId": "8F264O24",
      "quantity": 500000,
      "currentPrice": 150.30,
      "totalValue": 750905.00
    },
    {
      "stockSymbol": "TSLA",
      "stockName": "Tesla, Inc.",
      "stockExchange": "NASDAQ",
      "stockId": "HA266D40",
      "quantity": 20,
      "currentPrice": 720.50,
      "totalValue": 14410.00
    }
  ];

  constructor(private authService: AuthService) {}

  ngOnInit(): void {
    this.authService.getCurrentUser().subscribe(user => {
      this.username = user.username;
    });
    this.holdings.forEach(holding => {
      this.openStates.set(holding.stockId, false);
    });
  }

  onSelectStock(stock: any): void {
    if (this.selectedStock && this.selectedStock.stockId === stock.stockId) {
      this.selectedStock = null;
    } else {
      this.selectedStock = stock;
    }
  }

  toggleStock(stockId: string): void {
    const isOpen = this.openStates.get(stockId) || false;
    this.openStates.set(stockId, !isOpen);
  }
}
