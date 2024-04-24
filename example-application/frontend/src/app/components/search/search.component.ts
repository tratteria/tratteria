import { Component, OnDestroy, HostListener } from '@angular/core';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { StockService } from '../../services/stock.service';
import { Stock } from '../../models/stock.model';
import { SearchItem } from '../../models/search-item.model';
import { Router } from '@angular/router';

@Component({
  selector: 'app-search',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.css']
})
export class SearchComponent implements OnDestroy {
  searchResult: SearchItem[] = [];
  selectedStock: Stock | null = null;
  hasSearched: boolean = false;
  query: string = '';
  private destroy$ = new Subject<void>();

  lastMouseX = 0;
  lastMouseY = 0;

  constructor(
    private stockService: StockService,
    private router: Router
  ) { }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onSearch(event: Event): void {
    const target = event.target as HTMLInputElement;
    this.query = target.value.trim();
    this.hasSearched = this.query.length > 0;

    this.stockService.searchStocks(this.query)
      .pipe(takeUntil(this.destroy$))
      .subscribe((results) => {
        this.searchResult = results;
      });
  }

  onSelectStock(stockId: string): void {
    if (this.selectedStock && this.selectedStock.id === stockId) {
      this.selectedStock = null;
    } else {
      this.stockService.getStockDetails(stockId)
        .pipe(takeUntil(this.destroy$))
        .subscribe((stockDetails) => {
          this.selectedStock = stockDetails;
        });
    }
  }

  onBuyStock(stockId: string): void {
    this.router.navigate(['/order'], { queryParams: { action: 'Buy', stockId: stockId } });
  }
  
  onSellStock(stockId: string): void {
    this.router.navigate(['/order'], { queryParams: { action: 'Sell', stockId: stockId } });
  }

  @HostListener('document:mousemove', ['$event'])
  onMouseMove(e: MouseEvent) {
    this.lastMouseX = e.clientX;
    this.lastMouseY = e.clientY;
  }

  isMouseReallyOver(event: MouseEvent, searchItem: HTMLElement): boolean {
    const rect = searchItem.getBoundingClientRect();
    return (this.lastMouseX < rect.left || this.lastMouseX > rect.right ||
            this.lastMouseY < rect.top || this.lastMouseY > rect.bottom);
  }
}
