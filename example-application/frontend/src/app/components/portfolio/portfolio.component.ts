import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { StockService } from '../../services/stock.service';
import { Holding } from '../../models/holdings.model'
import { Router } from '@angular/router';

@Component({
  selector: 'app-portfolio',
  templateUrl: './portfolio.component.html',
  styleUrls: ['./portfolio.component.css']
})
export class PortfolioComponent implements OnInit {
  username: string = '';
  selectedStock: Holding | null = null;
  openStates: Map<string, boolean> = new Map();
  holdings: Holding[] = [];

  constructor(
    private authService: AuthService,
    private stockService: StockService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.fetchHoldings();
  }

  fetchHoldings(): void {
    this.stockService.getHoldings().subscribe(response => {
      if (response && response.holdings) {
        this.holdings = response.holdings;
        response.holdings.forEach(holding => {
          this.openStates.set(holding.stockID, false);
        });
      } else {
        this.holdings = [];
      }
    });
  }

  toggleStock(stockID: string): void {
    const isOpen = this.openStates.get(stockID) || false;
    this.openStates.set(stockID, !isOpen);
  }

  onBuyStock(stockId: string): void {
    this.router.navigate(['/order'], { queryParams: { action: 'Buy', stockId: stockId } });
  }
  
  onSellStock(stockId: string): void {
    this.router.navigate(['/order'], { queryParams: { action: 'Sell', stockId: stockId } });
  }
}
