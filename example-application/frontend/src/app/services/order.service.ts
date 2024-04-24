import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { TransactionDetails } from '../models/transaction-details.model';

@Injectable({
  providedIn: 'root'
})
export class OrderService {
  private orderAPIURL = environment.orderAPIURL;

  constructor(private http: HttpClient) { }

  placeOrder(stockId: string, orderType: string, quantity: number): Observable<TransactionDetails> {
    const orderDetails = { stockId, orderType, quantity };
    return this.http.post<TransactionDetails>(this.orderAPIURL, orderDetails);
  }

  getTransactionDetails(transactionId: string): Observable<TransactionDetails> {
    return this.http.get<TransactionDetails>(`${this.orderAPIURL}/${transactionId}`);
  }
}
