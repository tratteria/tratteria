import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-auth',
  templateUrl: './auth.component.html',
  styleUrls: ['./auth.component.css']
})
export class AuthComponent implements OnInit {
  isLoading: boolean = false;

  constructor(
    private authService: AuthService, 
    private router: Router,
    private activatedRoute: ActivatedRoute
  ) {}

  ngOnInit(): void {
    this.activatedRoute.queryParams.subscribe(params => {
      const code = params['code'];
      if (code) {
        this.isLoading = true;
        this.authService.exchangeCode(code).subscribe({
          next: () => {
            this.isLoading = false;
            this.router.navigate(['/']); 
          },
          error: (error) => {
            this.isLoading = false;
            console.error('Error exchanging code for token:', error);
            this.router.navigate(['/']); 
          }
        });
      } else {
        console.error('Code missing in the authentication callback.');
        this.router.navigate(['/']);
      }
    });
  }
  
  login(): void {
      this.authService.loginWithDex();
  }
}
