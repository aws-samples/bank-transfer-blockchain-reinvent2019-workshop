package bank

import (
	"encoding/json"
	"errors"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"github.com/shopspring/decimal"
)

type transferEvent struct {
	FromAccNumber string `json:"FromAccNumber"`
	FromBankID    string `json:"FromBankID"`
	ToAccNumber   string `json:"ToAccNumber"`
	ToBankID      string `json:"ToBankID"`
	Amount        string `json:"Amount"`
}

// Transfer funds from one account to another given four arguments: Payers account Id, Payees bank,
// Payees account Id, and the amount to transfer. The amount must be a positive number.
// If the payer and payee accounts belong to the same bank this will perform an intrabank transfer
// otherwise this will initiate an interbank transfer
// Transferring between accounts at the same bank, but with different currencies requires a Forex contract
// params: fromAccount, toBank, toAccount, amount
func (s *BankChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorret number of args. Expecting 4: fromAccount, toBank, toAccount, amount")
	}

	//sorting arguments
	fromAccNum := args[0]
	toBankID := args[1]
	toAccNum := args[2]
	amountAsString := args[3]
	amount, err := decimal.NewFromString(args[3])

	//validate amount is actually a number
	if err != nil {
		return shim.Error("Unable to parse amount: " + amountAsString)
	}

	if amount.LessThanOrEqual(decimal.NewFromFloat(0.0)) {
		return shim.Error("Second argument (Amount to deposit) must be a positive number")
	}

	//get the fromAccount
	fromAccountAsByes := s.queryAccount(stub, []string{fromAccNum}).Payload
	fromAccount := &account{}
	err = json.Unmarshal(fromAccountAsByes, fromAccount)

	if err != nil {
		return shim.Error("Unable to retrieve to account ID from ledger " + err.Error())
	}

	// check if funds are available
	if fromAccount.Balance.Cmp(amount) == -1 {
		return shim.Error("Account has insufficient funds")
	}

	//query our bank to get the ID of the institution
	bankAsBytes, err := stub.GetState("bank")

	if err != nil {
		return shim.Error("Unable to retrieve bank ID from ledger " + err.Error())
	}

	thisBank := &bank{}
	err = json.Unmarshal(bankAsBytes, thisBank)
	if err != nil {
		return shim.Error("Unable to retrieve bank ID from ledger " + err.Error())
	}

	// is this an interbank transfer? check if recipient bank is not this bank
	if toBankID != thisBank.ID {
		return shim.Error("Interbank transfer is not yet implemented, either implement it or use the solution")
	}

	//if not inter bank transfer, perform an intra bank transfer
	toAccountAsByes := s.queryAccount(stub, []string{toAccNum}).Payload
	toAccount := &account{}
	err = json.Unmarshal(toAccountAsByes, toAccount)

	if err != nil {
		return shim.Error("Unable to retrieve to account ID from ledger " + err.Error() + string(toAccountAsByes))
	}

	//check if from and to accounts use the same currency
	var exchangeRate decimal.Decimal
	if fromAccount.Currency == toAccount.Currency {
		exchangeRate = decimal.NewFromFloat(1.0)
	} else {
		// call handler function to invoke Forex chaincode
		exchangeRateAsFloat, err := getCurrencyConversion(stub, thisBank.ForexContract, fromAccount.Currency, toAccount.Currency)

		if err != nil {
			return shim.Error("Unable to perform currency conversion:" + err.Error())
		}

		exchangeRate = decimal.NewFromFloat(exchangeRateAsFloat)
	}

	//update balances
	fromAccount.Balance = fromAccount.Balance.Sub(amount)
	toAccount.Balance = toAccount.Balance.Add(amount.Mul(exchangeRate))

	// write changes to ledger
	fromAccountAsByes, _ = json.Marshal(fromAccount)
	err = stub.PutState(fromAccNum, fromAccountAsByes)
	if err != nil {
		return shim.Error("Error trying to commit account to ledger" + err.Error())
	}

	toAccountAsByes, _ = json.Marshal(toAccount)
	err = stub.PutState(toAccNum, toAccountAsByes)
	if err != nil {
		return shim.Error("Error trying to commit account to ledger" + err.Error())
	}

	//write out an event of the transfer
	event := &transferEvent{FromAccNumber: fromAccount.AccNumber, FromBankID: thisBank.ID, ToBankID: toBankID, ToAccNumber: toAccNum, Amount: amount.String()}
	eventBytes, _ := json.Marshal(event)
	stub.SetEvent("transfer-event", eventBytes)
	return (shim.Success(nil))
}

// }

func getCurrencyConversion(stub shim.ChaincodeStubInterface, forexContract string, baseCurrency string, counterCurrency string) (float64, error) {

	if forexContract == "" {
		return 0.0, errors.New("Forex contract is empty, unable to complete transaction")
	}

	// invoke the forex contract to get the exchange rate for the pair
	stringArgs := []string{"getForexPair", baseCurrency, counterCurrency}
	args := util.ArrayToChaincodeArgs(stringArgs)
	response := stub.InvokeChaincode(forexContract, args, "")

	if response.Status != shim.OK {
		return 0.0, errors.New("Unable to get exchange rate from Forex Contract" + response.Message)
	}

	responseForex := &forexPair{}
	err := json.Unmarshal(response.GetPayload(), responseForex)

	if err != nil {
		return 0.0, errors.New("Unable to unmarshal exchange rate from Forex Contract" + err.Error())
	}

	return responseForex.Rate, nil
}
