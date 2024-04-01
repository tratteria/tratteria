import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { Stock } from '../models/stock.model';
import { SearchItem } from '../models/search-item.model';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class StockService {
  private stockAPIURL = environment.stockAPIURL

  constructor(private http: HttpClient) { }

  searchStocks(query: string): Observable<SearchItem[]> {
    if (!query.trim()) {
      return of([]);
    }

    return this.http.get<{success: boolean, results: SearchItem[]}>(`${this.stockAPIURL}/search`, { params: { query }}).pipe(
      map(response => {
        if (response.success && response.results) {
          return response.results;
        }
        console.error('Received error response from server on stocks search:', response);
        return [];
      }),
      catchError(error => {
        console.error('Error searching stocks:', error);
        return of([]);
      })
    );
  }

  getStockDetails(stockId: string): Observable<Stock | null> {
    return this.http.get<Stock>(`${this.stockAPIURL}/${stockId}`).pipe(
      map(response => {
        return response;
      }),
      catchError(error => {
        console.error('Error fetching stock details:', error);
        return of(null);
      })
    );
  }
}
