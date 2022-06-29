import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Candidate } from '../models/candidate';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { ElectionResult } from '../models/election-result';
@Injectable({
  providedIn: 'root'
})
export class VoteService {

  constructor(private httpClient: HttpClient) { }

  listCandidate(): Observable<Candidate[]>{
    const result = this.httpClient.get<Candidate[]>(environment.baseUrl +"/api/candidates?page_id=1&page_size=10")
    return result
  }

  voteCandidate(nationalId: string, candidateId: string){
    return this.httpClient.post(environment.baseUrl +"/api/vote", { nationalId: nationalId,candidateId: candidateId});
  }

  voteStatus(nationalId: string){
    return this.httpClient.post(environment.baseUrl +"/api/vote/status", { nationalId: nationalId});
  }

  electionResult(){
    return this.httpClient.get<ElectionResult[]>(environment.baseUrl +"/election/result");
  }

  electionExport(){
    return this.httpClient.head<Blob>(environment.baseUrl +"election/export");
  }


}
