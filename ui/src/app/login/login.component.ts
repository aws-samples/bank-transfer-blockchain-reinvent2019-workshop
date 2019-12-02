import {Component} from '@angular/core';
import {Router} from '@angular/router';
import { environment } from 'src/environments/environment';
import { AccountService } from '../services/account.service';


@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent {

  accNum = '';
  bank_name = environment.bank_name;
  message = ''

  constructor(private router: Router, private accountService: AccountService) {
  }

  submit() {
console.log(this.accNum)
this.accountService.getAccount(this.accNum)
	  .subscribe((data)=>{
      console.log(data)
      this.message = ''
      this.router.navigateByUrl('/home', {state: {accNum: this.accNum}});
    },
    data => {
      console.log(data)
      this.message = "Unable to load account. Please check your configuration and ensure that the account exists"

  });
  }

}