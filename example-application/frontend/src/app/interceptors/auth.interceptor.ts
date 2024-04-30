import { Injectable } from '@angular/core';
import {
  HttpInterceptor,
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpErrorResponse
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Router } from '@angular/router';
import { AuthService } from '../services/auth.service';
import { ModalService } from '../services/modal.service';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(
    private router: Router,
    private authService: AuthService,
    private modalService: ModalService
  ) {}

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(req).pipe(
      catchError((error: HttpErrorResponse) => {
        if (error.status === 403) {
          console.log('Access forbidden. Showing modal...');
          this.modalService.open('Access Forbidden');
        } else if (error.status === 401) {
          console.log('Unauthorized response received from the server. Logging user out...');
          this.authService.logout().subscribe({
            next: () => {
              console.log('Logout process completed');
              this.router.navigate(['']);
            },
            error: (err) => {
              console.error('Error logging user out:', err);
              this.router.navigate(['']);
            }
          });
        }
        return throwError(error);
      })
    );
  }
}
