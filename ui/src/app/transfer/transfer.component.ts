import { Component, OnInit, Input} from '@angular/core';
import { AccountService } from '../services/account.service';


@Component({
  selector: 'app-transfer',
  templateUrl: './transfer.component.html',
  styleUrls: ['./transfer.component.css']
})
export class TransferComponent implements OnInit {

  constructor(private accountService: AccountService) { }

  amount;
  toAccNum;
  toBankID;
  message; 
  status;
  @Input() fromAccNum: string;
  show: boolean;

  ngOnInit() {
    this.show = true;
  }

  submit() {
    
    this.accountService.postTransfer(this.fromAccNum, this.toBankID, this.toAccNum, this.amount).subscribe(res => { 
      console.log(res);		
      this.message='Transfer Complete'
      this.show=false
      this.status='ok'

    },
    err => {
      console.log(err);
      this.message='Unable to make tranfer. Check the API configuraiton and ensure the tranfer details are correct.'
      this.show=false
      this.status='error'

    }
   );
  }

}
