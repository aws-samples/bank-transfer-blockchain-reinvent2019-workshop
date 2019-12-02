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

// deposit adds funds to an account, this is used to receive funds from an interbank transfer
//args
// 	acc 	string 	the account number to deposit funds to
// 	amount 	string	the amount to deposit
func (s *BankChaincode) deposit(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	accNum := args[0]
	amount, err := decimal.NewFromString(args[1])

	if err != nil {
		return shim.Error("Second argument (Amount to deposit) must be a number. Error: " + err.Error())
	}

	if amount.LessThanOrEqual(decimal.NewFromFloat(0.0)) {
		return shim.Error("Second argument (Amount to deposit) must be a positive number")
	}

	//get the account to operate on
	accountAsByes := s.queryAccount(stub, []string{accNum}).Payload
	acc := &account{}
	err = json.Unmarshal(accountAsByes, acc)

	if err != nil {
		return shim.Error("Unable to retrieve to account ID from ledger " + err.Error())
	}

	acc.Balance = acc.Balance.Add(amount)

	// write changes to ledger
	accAsBytes, _ := json.Marshal(acc)
	err = stub.PutState(acc.AccNumber, accAsBytes)
	if err != nil {
		return shim.Error("Error trying to commit account to ledger" + err.Error())
	}

	return (shim.Success(nil))

}
