import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { StockService } from '../../services/stock.service';
import { Holding } from '../../models/holdings.model';
import { Router } from '@angular/router';
import { CONSTANTS } from '../../config/constants';

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
  errorFetchingHoldings: boolean = false;
  constants = CONSTANTS;

  constructor(
    private authService: AuthService,
    private stockService: StockService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.fetchHoldings();
  }

  fetchHoldings(): void {
    this.stockService.getHoldings().subscribe({
      next: (response) => {
        if (response && response.holdings && response.holdings.length > 0) {
          this.holdings = response.holdings;
          response.holdings.forEach(holding => {
            this.openStates.set(holding.stockID, false);
          });
          this.errorFetchingHoldings = false;
        } else {
          this.holdings = [];
          this.errorFetchingHoldings = false;
        }
      },
      error: (error) => {
        console.error('Error fetching holdings:', error);
        this.errorFetchingHoldings = true;
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
