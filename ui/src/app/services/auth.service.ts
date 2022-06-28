import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import * as moment from "moment";
import { LoginResponse } from '../models/login-response';
import { tap, shareReplay } from 'rxjs/operators';
import { environment } from 'src/environments/environment';
import { UserService } from './user.service';


@Injectable({
  providedIn: 'root'
})
export class AuthService {

    constructor(private http: HttpClient, private userService: UserService) {}

    login(nationalId:string, password:string) {
        return this.http.post<LoginResponse>(environment.baseUrl + '/users/login', { national_id: nationalId, password})
        .pipe(
          tap(reult => this.setSession(reult)),
          shareReplay()
        );
    }

    private setSession(authResult: LoginResponse) {
        const expiresAt = moment(authResult.expired_at).valueOf();
        this.userService.setUser(authResult.user)
        localStorage.setItem('token', authResult.access_token);
        localStorage.setItem("expires_at", JSON.stringify(expiresAt));
    } 
          
    logout() {
        localStorage.removeItem("token");
        localStorage.removeItem("expires_at");
    }

    public isLoggedIn() {
        const token = localStorage.getItem("token");
        if (token == null) return false;

        return moment().isBefore(this.getExpiration());
    }

    isLoggedOut() {
        return !this.isLoggedIn();
    }

    getExpiration() {
        const expiration = localStorage.getItem("expires_at");
        if (expiration == null) return
        const expiresAt = JSON.parse(expiration);
        return moment(expiresAt);
    }    



}
          
          