import { Component, OnInit } from '@angular/core';
import { from, Observable } from 'rxjs';
import { Candidate } from '../models/candidate';
import { VoteService } from '../services/vote.service';

@Component({
  selector: 'app-vote',
  templateUrl: './vote.component.html',
  styleUrls: ['./vote.component.scss']
})
export class VoteComponent implements OnInit {

  constructor(private voteService: VoteService) { }

  candidates : Observable<Candidate[]> = from([])

  ngOnInit(): void {
    this.candidates = this.voteService.listCandidate();
  }

  refreshCandidates(){
    this.candidates = this.voteService.listCandidate();
  }

}
