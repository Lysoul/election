import { Component, OnInit } from '@angular/core';
import { map, Observable, of } from 'rxjs';
import { environment } from 'src/environments/environment';
import { AuthService } from '../../services/auth.service';
import { UserService } from '../../services/user.service';
import { VoteService } from '../../services/vote.service';

@Component({
  selector: 'app-nav',
  templateUrl: './nav.component.html',
  styleUrls: ['./nav.component.scss']
})
export class NavComponent implements OnInit {

  constructor(private authService: AuthService, 
    private userService: UserService,
    ) { }

  fullName: Observable<string> = of("Anonymous");
  hasLogedIn: Observable<boolean> = of(false);

  ngOnInit(): void {
    this.fullName = this.userService.getUser().pipe(map(x => x.full_name));
    this.hasLogedIn = this.userService.getUser().pipe(
      map(x => {
        if(x.national_id == "-1" || x.full_name == "Anonymous") return false;
        return true
      })
    );
  }

  logout(){
    this.authService.logout();
    this.refresh()
  }

  refresh(): void {
    window.location.reload();
  }


}
