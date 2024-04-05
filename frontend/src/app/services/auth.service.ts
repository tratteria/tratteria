import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, catchError, tap, Observable, throwError} from 'rxjs';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  
  private logoutAPIURL = environment.logoutAPIURL;
  private tokenExchangeAPIURL = environment.codeExchangeAPIURL;
  private dexHostURL = environment.dexHostURL
  private dexClientId = environment.dexClientId
  private isAuthenticated = new BehaviorSubject<boolean>(this.isLoggedIn());

  constructor(private http: HttpClient) { }

  get authState(): Observable<boolean> {
    return this.isAuthenticated.asObservable();
  }

  loginWithDex(): void {
    const clientId = this.dexClientId;
    const redirectUri = encodeURIComponent(window.location.origin + '/callback');
    const responseType = 'code';
    const scope = encodeURIComponent('openid profile email');

    window.location.href = `${this.dexHostURL}/dex/auth?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=${responseType}&scope=${scope}`;
  }

  logout(): Observable<any> {
    return this.http.post<any>(this.logoutAPIURL, {}).pipe(
      tap(() => {
        console.log('Logout successful');
        localStorage.removeItem('isLoggedIn');
        this.isAuthenticated.next(false);
      }),
      catchError(error => {
        console.error('Logout failed', error.message);
        return throwError(() => new Error('Logout failed. Please try again later.'));
      })
    );
  }

  exchangeCode(code: string): Observable<any> {
    return this.http.post<any>(this.tokenExchangeAPIURL, { code }).pipe(
      tap(() => {
        console.log('Code exchange successful.');
        localStorage.setItem('isLoggedIn', 'true');
        this.isAuthenticated.next(true);
      }),
      catchError(error => {
        console.error('Code exchange failed:', error.message);
        localStorage.removeItem('isLoggedIn');
        this.isAuthenticated.next(false);
        return throwError(() => new Error('Code exchange failed. Please try again later.'));
      })
    );
  }

  isLoggedIn(): boolean {
    return localStorage.getItem('isLoggedIn') === 'true';
  }
}
