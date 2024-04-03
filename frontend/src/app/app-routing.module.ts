import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { AuthComponent } from './components/auth/auth.component';
import { SearchComponent } from './components/search/search.component';
import { OrderComponent } from './components/order/order.component';
import { AuthGuard } from './guards/auth.guard';
import { TransactionDetailsComponent } from './components/order/transaction-details/transaction-details.component';
import { PortfolioComponent } from './components/portfolio/portfolio.component';
import { HomeComponent } from './components/home-component/home-component.component';


const routes: Routes = [
  { path: '', component: HomeComponent},
  { path: 'order', component: OrderComponent, canActivate: [AuthGuard] },
  { path: 'order/transaction', component: TransactionDetailsComponent, canActivate: [AuthGuard] },
  { path: 'portfolio', component: PortfolioComponent, canActivate: [AuthGuard] },
  { path: '**', redirectTo: '' }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
