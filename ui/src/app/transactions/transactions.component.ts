import { Component, Input, OnInit } from '@angular/core';
import { AccountService } from '../services/account.service';

@Component({
  selector: 'app-transactions',
  templateUrl: './transactions.component.html',
  styleUrls: ['./transactions.component.css']
})



export class TransactionsComponent implements OnInit {

constructor(private accountService: AccountService) { }
  @Input() accNum: string;
	transactions: any[];
	
  ngOnInit() {
        return this.accountService.getTransactions(this.accNum)
        .subscribe((data)=>{
          console.log(Array.isArray(data));
          
          var res = [];
          for (var x in data){
            data[x].Value = JSON.parse(data[x].Value);
            data[x].Timestamp = new Date(data[x].Timestamp*1000);
             res.push(data[x])
             console.log(data[x])
          }

          this.transactions = res.reverse();       
    });
  }

}