/*
# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License").
# You may not use this file except in compliance with the License.
# A copy of the License is located at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
#
*/

package interbank

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"github.com/shopspring/decimal"
)

type route struct {
	ID            string `json:"name"`
	BankContract  string `json:"bankContract"`
	ForexContract string `json:"ForexContract"`
}

type account struct {
	Name      string          `json:"name"`
	AccNumber string          `json:"ID"`
	Balance   decimal.Decimal `json:"balance"`
	Currency  string          `json:"currency"`
}

type forexPair struct {
	Pair string  `json:"pair"`
	Rate float64 `json:"rate"`
}

// InterbankChaincode is the struct to which all contract methods are associated with
type InterbankChaincode struct {
}

// Initalize the chaincode
func (s *InterbankChaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

//Invoke is called when external applications invoke the smart contract
func (s *InterbankChaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "interbankTransfer" {
		return s.interbankTransfer(stub, args)
	} else if function == "registerRoute" {
		return s.registerRoute(stub, args)
	}

	return shim.Error("Invalid function")
}

// Perform a transfer between two banks
// params:
//	toAccNumber	string	the account number to pay
//	toBankID	string	the ID of the bank that the account belongs to
//	amount		string	the amount to pay
//	currency	string 	the currency of the amount being paid
func (s *InterbankChaincode) interbankTransfer(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	toAccNum := args[0]
	toBankID := args[1]
	amount := args[2]
	currency := args[3]

	routeAsBytes, err := stub.GetState(toBankID)

	if err != nil {
		return shim.Error(err.Error())
	}

	toRoute := &route{}
	err = json.Unmarshal(routeAsBytes, toRoute)
	if err != nil {
		return shim.Error(err.Error())
	}

	toBankContract := toRoute.BankContract

	stringArgs := []string{"queryAccount", toAccNum}
	response := stub.InvokeChaincode(toBankContract, util.ArrayToChaincodeArgs(stringArgs), "")
	toAccount := &account{}
	err = json.Unmarshal(response.Payload, toAccount)

	if err != nil {
		return shim.Error("Unable to retrieve to account from ledger " + err.Error())
	}

	var exchangeRate decimal.Decimal

	//check if currency conversion is required
	if currency == toAccount.Currency {
		exchangeRate = decimal.NewFromFloat(1.0)
	} else {
		forexContract := toRoute.ForexContract
		exchangeRateAsFloat, err := currencyConversion(stub, forexContract, currency, toAccount.Currency)

		if err != nil {
			return shim.Error("Unable to perform currency conversion " + err.Error())
		}

		exchangeRate = decimal.NewFromFloat(exchangeRateAsFloat)
	}

	//perform payment
	amountAsDecimal, err := decimal.NewFromString(amount)

	if err != nil {
		return shim.Error(err.Error())
	}

	amountAsDecimal = amountAsDecimal.Mul(exchangeRate)
	amountAsString := amountAsDecimal.String()

	stringArgs = []string{"deposit", toAccNum, amountAsString}
	response = stub.InvokeChaincode(toBankContract, util.ArrayToChaincodeArgs(stringArgs), "")

	if response.GetStatus() != shim.OK {
		return shim.Error("Failed to make payment " + err.Error())
	}

	return (shim.Success(nil))
}

func currencyConversion(stub shim.ChaincodeStubInterface, forexContract string, baseCurrency string, counterCurrency string) (float64, error) {

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

//registerRoute registers the contracts associated with a bank to allow for interbank transfer
//Args
//	ID            string 	the ID of the bank, analogue of SWIFT, IBAN, routing code, etc
//	BankContract  string 	the name of the chaincode for the bank
//	ForexContract string 	the name of the contract to provide forex services
func (s *InterbankChaincode) registerRoute(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Expecting 3 arguments: bank ID, name of BankChaincode, name of ForexContract")
	}

	newRoute := route{ID: args[0], BankContract: args[1], ForexContract: args[2]}
	asBytes, _ := json.Marshal(newRoute)
	err := stub.PutState(args[0], asBytes)

	if err != nil {
		return shim.Error("Unable to commit route to ledger")
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(InterbankChaincode))
	if err != nil {
		fmt.Printf("Error creating new InterbankChaincodet: %s", err)
	}

}
