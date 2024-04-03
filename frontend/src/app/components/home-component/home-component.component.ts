import { Component, OnDestroy } from '@angular/core';
import { AuthService } from '../../services/auth.service';
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-home',
  template: `
    <app-auth *ngIf="!isAuthenticated"></app-auth>
    <app-search *ngIf="isAuthenticated"></app-search>
  `,
})
export class HomeComponent implements OnDestroy {
  isAuthenticated: boolean = this.authService.isLoggedIn();
  private authSubscription: Subscription;

  constructor(private authService: AuthService) {
    this.authSubscription = this.authService.authState.subscribe(
      (isAuthenticated) => {
        this.isAuthenticated = isAuthenticated;
      }
    );
  }

  ngOnDestroy(): void {
    this.authSubscription.unsubscribe();
  }
}
