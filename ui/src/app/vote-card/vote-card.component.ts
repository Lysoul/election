import { Component, OnInit, ChangeDetectionStrategy, Input, OnDestroy, Output, EventEmitter } from '@angular/core';
import { ToastrService } from 'ngx-toastr';
import { Subscription } from 'rxjs';
import { Candidate } from '../models/candidate';
import { UserService } from '../services/user.service';
import { VoteService } from '../services/vote.service';
import { mergeMap, switchMap } from 'rxjs/operators';

@Component({
  selector: 'app-vote-card',
  templateUrl: './vote-card.component.html',
  styleUrls: ['./vote-card.component.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class VoteCardComponent implements OnInit, OnDestroy {

  constructor(private voteService: VoteService, 
    private userService: UserService,
    private toastr: ToastrService) { }
 
  ngOnDestroy(): void {
    this.voteSubscription.unsubscribe();
  }

  voteSubscription: Subscription = new Subscription();

  @Input() candidate: Candidate = {
      id: '-1',
      name: '',
      dob: '',
      bio_link: '',
      image_url: '',
      policy: '',
      vote_count: 0,
  };

  @Output() hasUpdeVote = new EventEmitter<void>();

  ngOnInit(): void {
  }

  vote() {
    const user = this.userService.getCurrentUser();

    this.voteSubscription = this.voteService.voteCandidate(user.national_id, this.candidate.id)
    .subscribe({
      next: () =>{
        this.toastr.success('Success!', `You are voted ${this.candidate.name}!`,{
          timeOut: 2000,
        });
        this.hasUpdeVote.emit();
      },
      error: (err) => {
        this.toastr.error('Error!', `You are already voted!`,{
          timeOut: 2000,
        });
      },
    });

  } 

}
