'use strict';

import { h, Component } from 'preact';
import { bind } from 'decko';
import Pikadayer from './Pikadayer.js';
import classNames from 'classnames';

const padZero = (n) => n < 10 ? '0'+n : ''+n;

let now = new Date();
window.setInterval(() => {
  now = new Date();
}, 60000 );

const availablePeriods = {
  '1d': {
    label: '1d',
    start: function() {
      return new Date(now.getFullYear(), now.getMonth(), now.getDate());
    },
    end: function() {
      return this.start();
    },
  },
  '1w': {
    label: '1w',
    start: function() {
      return new Date(now.getFullYear(), now.getMonth(), now.getDate()-6);
    },
    end: function() {
      return new Date(now.getFullYear(), now.getMonth(), now.getDate());
    },
 },
 '4w': {
    label: '4w',
    start: function() {
      return new Date(now.getFullYear(), now.getMonth(), now.getDate()-4*7+1);
    },
    end: function() {
      return new Date(now.getFullYear(), now.getMonth(), now.getDate());
    },
 },
 'mtd': {
    label: 'Mtd',
    start: function() {
      return new Date(now.getFullYear(),  now.getMonth(), 1);
    },
    end: function() {
      return new Date(now.getFullYear(), now.getMonth()+1, 0);
    },
 },
'qtd': {
  label: 'Qtd',
  start: function() {
    let qs = Math.ceil((now.getMonth()+1) / 3) * 3 - 3;
    return new Date(now.getFullYear(), qs, 1);

  },
  end: function() {
    let start = this.start();
    return new Date(start.getFullYear(), start.getMonth() + 3, 0);
  },
 },
 'ytd': {
  label: 'Ytd',
  start: function() {
    return new Date(now.getFullYear(), 0, 1);
  },
  end: function() {
    return new Date(now.getFullYear()+1, 0, 0);
  },
 },
 'all': {
  label: 'All',
  start: function() {
    return new Date(2018, 6, 1);
  },
  end: function() {
    return new Date();
  },
 }
}

function hashParams() {
 var params = {}, 
  match, 
  matches =  window.location.hash.substring(2).split("&");

 for(var i=0; i<matches.length; i++) {
   match = matches[i].split('=')
   params[match[0]] = decodeURIComponent(match[1]);
 }

 return params;
}

class DatePicker extends Component {
  constructor(props) {
    super(props)

    let params = hashParams();

    this.state = {
      period: params.p || window.localStorage.getItem('period') || '1w',
      startDate: new Date(params.s || 'now'),
      endDate: new Date(params.e || 'now'),
      groupBy: params.g || 'day',
    }    
    this.state.diff = this.calculateDiff(this.state.startDate, this.state.endDate)

    if(this.state.period !== 'custom') {
      this.updateDatesFromPeriod(this.state.period, params.g)
    } else {
      this.props.onChange(this.state);
    }
  }

  componentDidMount() {
    window.addEventListener('keydown', this.handleKeyPress);
  }

  componentWillUnmount() {
    window.removeEventListener('keydown', this.handleKeyPress)
  }

  @bind
  updateDatesFromPeriod(period, groupBy) {
    if(typeof(availablePeriods[period]) !== "object") {
      period = "1w";
    }
    let p = availablePeriods[period];
    this.setDateRange(p.start(), p.end(), period, groupBy);
  }

  @bind
  setDateRange(start, end, period, groupBy) {
    // don't update state if start > end. user may be busy picking dates.
    if(start > end) {
      return;
    }

    // include start & end day by forcing time
    start.setHours(0, 0, 0);
    end.setHours(23, 59, 59);

    let diff =  this.calculateDiff(start, end)
    if(!groupBy) {
      groupBy = 'day';

      if(diff >= 31) {
        groupBy = 'month';
      } else if( diff < 2) {
        groupBy = 'hour';
      }
    }
   
   
    this.setState({
      period: period,
      startDate: start,
      endDate: end,
      diff: diff,
      groupBy: groupBy,
    });

    // use slight delay for updating rest of application to allow this function to be called again
    if(!this.timeout) {
      this.timeout = window.setTimeout(() => {
        this.props.onChange(this.state);
        this.updateURL()
        this.timeout = null;
      }, 5)
    }
  }

  calculateDiff(start, end) {
    return Math.round((end - start) / 1000 / 60 / 60 / 24)
  }

  updateURL() {
    if(this.state.period !== 'custom') {
      window.history.replaceState(this.state, null, `#!p=${this.state.period}&g=${this.state.groupBy}`)
    } else {
      window.history.replaceState(this.state, null, `#!p=custom&s=${encodeURIComponent(this.state.startDate.toISOString())}&e=${encodeURIComponent(this.state.endDate.toISOString())}&g=${this.state.groupBy}`)
    }
  }

  @bind
  setPeriod(e) {
    e.preventDefault();

    let newPeriod = e.target.getAttribute('data-value');
    if( newPeriod === this.state.period) {
      return;
    }

    window.localStorage.setItem('period', this.state.period)
    this.updateDatesFromPeriod(newPeriod);
  }

  dateValue(date) {
    return date.getFullYear() + '-' + padZero(date.getMonth() + 1) + '-' + padZero(date.getDate());
  }

  @bind
  setStartDate(date) {
    this.setDateRange(date, this.state.endDate, 'custom')
  }

  @bind
  setEndDate(date) {
    this.setDateRange(this.state.startDate, date, 'custom')
  }

  @bind
  handleKeyPress(evt) {
    // Don't handle input when the user is in a text field or text area.
    let tag = evt.target.tagName;
    if(tag === "INPUT" || tag === "TEXTAREA") {
      return;
    }

    // TODO: Account for leap years
    let diff = this.state.endDate - this.state.startDate + 1000;
    let newStartDate, newEndDate;

    switch(evt.which) {
      // left-arrow
      case 37:
        newStartDate = new Date(+this.state.startDate - diff)
        newEndDate = new Date(+this.state.endDate - diff)
        this.setDateRange(newStartDate, newEndDate)
      break;

      //right-arrow
      case 39:
      newStartDate = new Date(+this.state.startDate + diff)
      newEndDate = new Date(+this.state.endDate + diff)
      this.setDateRange(newStartDate, newEndDate)
      break;
    }
  }

  @bind
  setGroupBy(e) {
    this.setState({
      groupBy: e.target.getAttribute('data-value')
    })
    this.props.onChange(this.state);
    this.updateURL()
  }

  render(props, state) {
    const presets = Object.keys(availablePeriods).map((id) => {
      let p = availablePeriods[id];
      return (
        <li class={classNames({ current: id == state.period })}>
          <a href="javascript:void(0);" data-value={id} onClick={this.setPeriod}>{p.label}</a>
        </li>
      );
    });

    return (
      <nav class="date-nav sm ac">
        <ul>
          {presets}
        </ul>
        <ul>
          <li><Pikadayer value={this.dateValue(state.startDate)} onSelect={this.setStartDate} /> <span>›</span> <Pikadayer value={this.dateValue(state.endDate)} onSelect={this.setEndDate}  /></li>
        </ul>
        <ul>
         {state.diff < 31 ? (<li class={classNames({ current: 'hour' === state.groupBy })}><a href="javascript:;" data-value="hour" onClick={this.setGroupBy}>Hourly</a></li>) : ''}
         <li class={classNames({ current: 'day' === state.groupBy })}><a href="javascript:;" data-value="day" onClick={this.setGroupBy}>Daily</a></li>
         {state.diff >= 31 ? (<li class={classNames({ current: 'month' === state.groupBy })}><a href="javascript:;" data-value="month" onClick={this.setGroupBy}>Monthly</a></li>) : ''}
        </ul>
      </nav>
    )

    /*
    <ul>
        <li class="current"><a href="#">Daily</a></li>
        <li><a href="#">Monthly</a></li>
    </ul>
    */

  }
}

export default DatePicker
