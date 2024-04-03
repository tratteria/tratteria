import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { OrderService } from '../../../services/order.service'
import { TransactionDetails } from '../../../models/transaction-details.model';

@Component({
  selector: 'app-transaction-details',
  templateUrl: './transaction-details.component.html',
  styleUrls: ['./transaction-details.component.css']
})
export class TransactionDetailsComponent implements OnInit {
  transactionId: string | null = null;
  transactionDetails: TransactionDetails | null = null;

  constructor(
    private activatedRoute: ActivatedRoute,
    private orderService: OrderService
  ) { }

  ngOnInit(): void {
    this.activatedRoute.queryParams.subscribe(params => {
      this.transactionId = params['transaction_id'];
      if (this.transactionId) {
        this.fetchTransactionDetails(this.transactionId);
      }
    });
  }

  fetchTransactionDetails(id: string): void {
    this.orderService.getTransactionDetails(id).subscribe({
      next: (details) => {
        this.transactionDetails = details;
      },
      error: (error) => {
        console.error('Error fetching transaction details:', error);
      }
    });
  }
}
