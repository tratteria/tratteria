import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, catchError, tap, Observable, throwError, of } from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  
  private loginAPIURL = environment.loginAPIURL;
  private logoutAPIURL = environment.logoutAPIURL;
  
  private isAuthenticated = new BehaviorSubject<boolean>(this.isLoggedIn());

  constructor(private http: HttpClient) { }

  get authState(): Observable<boolean> {
    return this.isAuthenticated.asObservable();
  }

  login(username: string): Observable<any> {
    return this.http.post<any>(this.loginAPIURL, { username }).pipe(
      tap(() => {
        console.log('Login successful');
        localStorage.setItem('isLoggedIn', 'true');
        localStorage.setItem('username', username);
        this.isAuthenticated.next(true);
      }),
      catchError(error => {
        console.error('Login failed', error.message);
        localStorage.removeItem('isLoggedIn');
        localStorage.removeItem('username');
        this.isAuthenticated.next(false);
        return throwError(() => new Error('Login failed. Please try again later.'));
      })
    );
  }

  logout(): Observable<any> {
    return this.http.post<any>(this.logoutAPIURL, {}).pipe(
      tap(() => {
        console.log('Logout successful');
        localStorage.removeItem('isLoggedIn');
        localStorage.removeItem('username');
        this.isAuthenticated.next(false);
      }),
      catchError(error => {
        console.error('Logout failed', error.message);
        return throwError(() => new Error('Logout failed. Please try again later.'));
      })
    );
  }
  
  isLoggedIn(): boolean {
    return localStorage.getItem('isLoggedIn') === 'true';
  }

  getCurrentUser(): Observable<{ username: string }> {
    const username = localStorage.getItem('username') || 'Unknown';
    return of({ username });
  }
}
