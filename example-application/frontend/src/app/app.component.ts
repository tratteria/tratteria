import { Component, OnDestroy } from '@angular/core';
import { Router, NavigationEnd } from '@angular/router';
import { filter, Subscription } from 'rxjs';
import { AuthService } from './services/auth.service';
import { CONSTANTS } from './config/constants';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnDestroy {
  showHomeButton: boolean = false;
  showProfileOptions: boolean = false;
  
  private authSubscription: Subscription;

  constructor(private authService: AuthService, private router: Router) {
    this.authSubscription = this.authService.authState.subscribe(isAuthenticated => {
      this.showProfileOptions = isAuthenticated;
    });

    this.router.events.pipe(
      filter((event): event is NavigationEnd => event instanceof NavigationEnd)
    ).subscribe((event: NavigationEnd) => {
      const isHomePage = event.urlAfterRedirects === '/';
      this.showHomeButton = !isHomePage;
    });
  }

  logout(): void {
    this.authService.logout().subscribe({
      next: () => {
        console.log('Logout successful');
        this.router.navigate(['']);
      },
      error: (error) => {
        console.error('Logout error', error);
      }
    });
  }

  ngOnDestroy(): void {
    this.authSubscription.unsubscribe();
  }
}
