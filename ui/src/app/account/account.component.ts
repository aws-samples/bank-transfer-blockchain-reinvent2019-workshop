import { Component, Input, OnInit } from '@angular/core';
import { AccountService } from '../services/account.service';

@Component({
  selector: 'app-account',
  templateUrl: './account.component.html',
  styleUrls: ['./account.component.css']
})
export class AccountComponent implements OnInit {
  @Input() accNum: string;
	account;

constructor(private accountService: AccountService) { }

  ngOnInit() {
	  this.account = this.accountService.getAccount(this.accNum)
	  .subscribe((data)=>{
      console.log(data);
      this.account = data;
    });
  }


}
