import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { StockService } from '../../services/stock.service';
import { Stock } from '../../models/stock.model';

@Component({
  selector: 'app-order',
  templateUrl: './order.component.html',
  styleUrls: ['./order.component.css']
})
export class OrderComponent implements OnInit {
  action: string = '';
  quantity: number = 1;
  total: number = 0;
  stock: Stock | null = null;

  constructor(
    private activatedRoute: ActivatedRoute,
    private stockService: StockService,
    private router: Router,
  ) { }

  ngOnInit(): void {
    this.activatedRoute.queryParams.subscribe(params => {
      const action = params['action'];
      const stockId = params['stockId'];
      if (action && stockId) {
        this.action = action;
        this.stockService.getStockDetails(stockId).subscribe(stockDetails => {
          this.stock = stockDetails;
          this.calculateTotal();
        });
      }
    });
  }

  calculateTotal(): void {
    if (this.stock) {
      const pricePerShare = parseFloat(this.stock.currentPrice);
      this.total = pricePerShare * this.quantity;
    }
  }

  fixNumber(event: any): void {
    if (event.target.value < 0 || event.target.value == 0) {
      event.target.value = '';
      this.quantity = 1
      this.calculateTotal()
    }
  }

  onFocus(event: any): void {
    if (event.target.value == '1') {
      event.target.value = '';
    }
  }
  
  placeOrder(): void {
    const mockTransactionId = Date.now();

    this.router.navigate(['/order/transaction'], { queryParams: { transaction_id: mockTransactionId } });
  }
}
