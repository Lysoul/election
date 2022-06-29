import { Component, OnInit } from '@angular/core';
import { ScaleType } from '@swimlane/ngx-charts';

import { Color } from '@swimlane/ngx-charts/lib/utils/color-sets';
import { map, Observable, of } from 'rxjs';
import { ChartElection } from '../../models/election-result';
import { VoteService } from '../../services/vote.service';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {

  //Chart Options
  showXAxis = true;
  showYAxis = true;
  gradient = false;
  showLegend = true;
  showXAxisLabel = true;
  xAxisLabel = 'Candidate';
  showYAxisLabel = true;
  yAxisLabel = 'Vote';
  lgTitle = "Candidates"

  electionResult: Observable<ChartElection[]> = of([]);
  view: [number,number] = [850, 550];

  colorScheme: Color = { 
    domain: ['#5AA454', '#E44D25', '#CFC0BB', '#7aa3e5', '#a8385d', '#aae3f5'],
    name: '',
    selectable: true,
    group: ScaleType.Linear
  }

  cardColor: string = '#232837';
  
  constructor(private voteService: VoteService) {

  }

  
  ngOnInit(): void {
    this.electionResult = this.voteService.electionResult().pipe(
      map(result => {
        return result.map(x => {
          const chart: ChartElection = {
            name: x.name,
            value: x.vote_count
          }
          return chart
      })
    }))
  }

  onSelect(event:any) {
    console.log(event);
  }


}
