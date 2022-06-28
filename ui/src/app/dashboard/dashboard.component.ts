import { Component, OnInit } from '@angular/core';
import { ScaleType } from '@swimlane/ngx-charts';

import { Color } from '@swimlane/ngx-charts/lib/utils/color-sets';
import { single } from './data';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {

  single: any[] = [];
  view: [number,number] = [850, 550];

  colorScheme: Color = { 
    domain: ['#5AA454', '#E44D25', '#CFC0BB', '#7aa3e5', '#a8385d', '#aae3f5'],
    name: '',
    selectable: true,
    group: ScaleType.Linear
  }

  cardColor: string = '#232837';
  
  constructor() {
    this.single = single
  }
  ngOnInit(): void {

  }

  onSelect(event:any) {
    console.log(event);
  }

  showXAxis = true;
  showYAxis = true;
  gradient = false;
  showLegend = true;
  showXAxisLabel = true;
  xAxisLabel = 'Candidate';
  showYAxisLabel = true;
  yAxisLabel = 'Vote';
  lgTitle = "Candidates"
}
