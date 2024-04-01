import { Component } from '@angular/core';
import { Router, NavigationEnd } from '@angular/router';
import { filter } from 'rxjs/operators';
import { AuthService } from './services/auth.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'Alpha Stocks | Revolutionizing Online Trading';
  showHomeButton: boolean = false;
  showProfileOptions: boolean = false;

  constructor(private authService: AuthService, private router: Router) {
    this.router.events.pipe(
      filter((event): event is NavigationEnd => event instanceof NavigationEnd)
    ).subscribe((event: NavigationEnd) => {
      const isSearchPage = event.urlAfterRedirects === '/search';
      const isAuthPage = event.urlAfterRedirects === '/auth';

      this.showHomeButton = !isSearchPage && !isAuthPage;

      this.showProfileOptions = !isAuthPage;
    });
  }

  logout(): void {
    this.authService.logout().subscribe({
      next: () => {
        console.log('Logout successful');
        this.router.navigate(['/auth']);
      },
      error: (error) => {
        console.error('Logout error', error);
      }
    });
  }
}
