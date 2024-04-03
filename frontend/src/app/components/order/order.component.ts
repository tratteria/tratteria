import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { StockService } from '../../services/stock.service';
import { Stock } from '../../models/stock.model';
import { OrderService } from '../../services/order.service';

@Component({
  selector: 'app-order',
  templateUrl: './order.component.html',
  styleUrls: ['./order.component.css']
})
export class OrderComponent implements OnInit {
  action: string = '';
  quantity: number = 1;
  maxQuantity: number = 100;
  total: number = 0;
  stock: Stock | null = null;

  constructor(
    private activatedRoute: ActivatedRoute,
    private stockService: StockService,
    private orderService: OrderService,
    private router: Router,
  ) { }

  ngOnInit(): void {
    this.activatedRoute.queryParams.subscribe(params => {
      const action = params['action'];
      const stockId = params['stockId'];
      if (action && stockId) {
        this.action = action;
        this.stockService.getStockDetails(stockId).subscribe(stockDetails => {
          if (!stockDetails) {
            console.error('Stock details not found for ID:', stockId);
            return;
          }
          this.stock = stockDetails;
          this.maxQuantity = this.action === 'Buy' ? this.stock.totalAvailableShares : this.stock.holdings;
          this.calculateTotal();
        });
      }
    });
  }

  calculateTotal(): void {
    if (this.stock) {
      this.total = this.stock.currentPrice * this.quantity;
    }
  }

  fixNumber(event: any): void {
    if (event.target.value < 0 || event.target.value == 0) {
      event.target.value = '';
      this.quantity = 1
      this.calculateTotal()
    }
    if (event.target.value > this.maxQuantity) {
      event.target.value = this.maxQuantity;
      this.quantity = this.maxQuantity
      this.calculateTotal()
    }
  }

  onFocus(event: any): void {
    if (event.target.value == '1') {
      event.target.value = '';
    }
  }
  
  placeOrder(): void {
    if (this.stock && this.stock.id && this.action) {
      this.orderService.placeOrder(this.stock.id, this.action, this.quantity).subscribe({
        next: (response) => {
          this.router.navigate(['/order/transaction'], { queryParams: { transaction_id: response.transactionID } });
        },
        error: (error) => {
          console.error('Error placing order:', error);
        }
      });
    }
  }
}
