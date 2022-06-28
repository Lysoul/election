import { Component, OnDestroy, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Subscription } from 'rxjs';
import { AuthService } from '../services/auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit, OnDestroy {

  constructor(private authService: AuthService,
    private router: Router) { }

  nationalid = "";
  password = "";

  ngOnInit(): void {
  }

  logInSubscription: Subscription = new Subscription()

  login() {
    this.logInSubscription = this.authService.login(this.nationalid,this.password)
    .subscribe({
      next: () =>{
        this.router.navigate(['/']);
      }
    });
    
  }


  ngOnDestroy(): void {
    this.logInSubscription.unsubscribe();
  }


}
