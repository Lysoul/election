import { Pipe, PipeTransform } from '@angular/core';
import * as moment from 'moment';

@Pipe({
  name: 'age'
})
export class AgePipe implements PipeTransform {

  transform(value: string): string {
    let today = moment();
    let birthdate = moment(value,'MMMM DD, YYYY');
    let years = today.diff(birthdate, 'years');
    return `${isNaN(years) ? 0 : years} years`;
  }

}
