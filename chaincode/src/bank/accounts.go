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

package bank

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"github.com/shopspring/decimal"
)

// createAccount creates a new bank account at this bank
//Args:
//	Name      string          The customer name
//	AccNumber string          The account number
//	Balance   decimal.Decimal The account balance
//	Currency  string          The three decimal currency code for the account

func (s *BankChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect arguments, expecting customer name, account number, balance and, currency")
	}

	balance, decimalError := decimal.NewFromString(args[2])

	if decimalError != nil {
		return shim.Error("Unable to parse account balance")
	}

	account := account{Name: args[0], AccNumber: args[1], Balance: balance, Currency: args[3]}

	//serialize account
	accountBytes, _ := json.Marshal(account)

	//Add account to ledger
	putStateErr := stub.PutState(args[1], accountBytes)

	if putStateErr != nil {
		return shim.Error("Failed to create bank")
	}

	return shim.Success(nil)

}

// queryAccount returns a record for an account
//Args:
//	AccNumber string          The account number
func (s *BankChaincode) queryAccount(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting the account number")
	}

	accountAsBytes, stubError := stub.GetState(args[0])

	if stubError != nil {
		return shim.Error(stubError.Error())

	}

	return (shim.Success(accountAsBytes))
}
