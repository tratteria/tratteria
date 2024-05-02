import { Component, OnDestroy, HostListener } from '@angular/core';
import { FormControl } from '@angular/forms';
import { Subject, auditTime, takeUntil } from 'rxjs';
import { StockService } from '../../services/stock.service';
import { Stock } from '../../models/stock.model';
import { SearchItem } from '../../models/search-item.model';
import { Router } from '@angular/router';

const AUDIT_INTERVAL_MS = 500;

@Component({
  selector: 'app-search',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.css']
})
export class SearchComponent implements OnDestroy {
  searchResult: SearchItem[] = [];
  selectedStock: Stock | null = null;
  hasSearched: boolean = false;
  searchInput = new FormControl();
  private destroy$ = new Subject<void>();
  query: string = '';

  lastMouseX = 0;
  lastMouseY = 0;

  constructor(private stockService: StockService, private router: Router) {
    this.searchInput.valueChanges.pipe(
      auditTime(AUDIT_INTERVAL_MS),
      takeUntil(this.destroy$)
    ).subscribe(query => {
      this.query = query.trim();
      this.hasSearched = this.query.length > 0;
      
      if (this.hasSearched) {
        this.stockService.searchStocks(this.query).subscribe(results => {
          this.searchResult = results;
        });
      } else {
        this.searchResult = [];
      }
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  onSelectStock(stockId: string): void {
    if (this.selectedStock && this.selectedStock.id === stockId) {
      this.selectedStock = null;
    } else {
      this.stockService.getStockDetails(stockId)
        .pipe(takeUntil(this.destroy$))
        .subscribe(stockDetails => {
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
  onMouseMove(e: MouseEvent): void {
    this.lastMouseX = e.clientX;
    this.lastMouseY = e.clientY;
  }

  isMouseReallyOver(event: MouseEvent, searchItem: HTMLElement): boolean {
    const rect = searchItem.getBoundingClientRect();
    return (this.lastMouseX < rect.left || this.lastMouseX > rect.right ||
            this.lastMouseY < rect.top || this.lastMouseY > rect.bottom);
  }
}
