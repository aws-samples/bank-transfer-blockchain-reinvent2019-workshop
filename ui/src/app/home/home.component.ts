import { Component, OnInit } from '@angular/core';
import { environment } from 'src/environments/environment';
import {Router} from '@angular/router';


@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit {

  transferVisible = false;
  bank_name = environment.bank_name;
  accNum;

  constructor(public router: Router) { 
    this.accNum = this.router.getCurrentNavigation().extras.state.accNum;
  }

  ngOnInit() {
    
  }

public showTransfer(){
  this.transferVisible = !this.transferVisible
}
}
