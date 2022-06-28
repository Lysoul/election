import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable, take  } from 'rxjs';
import { DefaultUser, User } from '../models/user';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  private $currentUser: BehaviorSubject<User> = new BehaviorSubject<User>(DefaultUser)

  constructor() { }

  setUser(user: User){
    this.$currentUser.next(user)
  }

  getUser(): Observable<User>{
    return this.$currentUser.asObservable();
  }

  getCurrentUser(): User{
    return this.$currentUser.value;
  }


  destroyUser(){
    this.$currentUser.complete();
  }
}
