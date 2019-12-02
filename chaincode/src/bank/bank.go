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

//BankChaincode is the struct that all chaincode methods are associated with
//The bank chaincode provides a simple representation of a bank. It allows for the creation of bank
//accounts and the transfer of funds between bank accounts. There are several functions:
//	init - initialize the chaincode
//	invoke - called upon invocation and calls other functions
//	createAccount - create a bank account
//	deposit - deposit funds into a bank account
//	transfer - transfer funds between accounts, either interbank or intrabank
type BankChaincode struct {
}

// bank is a struct that represents a bank, it is stored on the ledeger under the key "bank"
//Name string - the name of the bank
//ID string - the ID of the bank, used to route between banks, analogous to an IBAN or SWIFT code
//ForexContract - the name of a ForexChaincode, deployed to the same peer as the bank, use to provide intrabank currency exchange
//InterbankContract - the name of the InterbankChaincode, used to transfer funds between banks
// A bank contract must be initalized with and name and ID. The two contracts are optional but required to do interbank transfers
//and interbank currency exchange - without them these will produce an error.
type bank struct {
	Name              string `json:"name"`
	ID                string `json:"bankID"`
	ForexContract     string `json:"forexContract"`
	InterbankContract string `json:"interbankContract"`
}

type account struct {
	Name      string          `json:"name"`
	AccNumber string          `json:"id"`
	Balance   decimal.Decimal `json:"balance"`
	Currency  string          `json:"currency"`
}

type forexPair struct {
	Pair string  `json:"pair"`
	Rate float64 `json:"rate"`
}

//Init method is run on chaincode installation and upgrade
//Args:
// 	Name 				string 		The Name of the Bank
//	ID					string		The institution ID of the bank, (e.g. IBAN, SWIFT or other routing code)
//	ForexContract		string		The name of the contract that provides Forex services to this bank
//  InterbankContract	string	The name of the contrat providing interbank transfer to this bank
func (s *BankChaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	args := stub.GetStringArgs()
	if len(args) < 2 {
		return shim.Error("Incorrect arguments. Expecting a bank name, ID. Optionally and the name of the ForexContract and the name of the InterBank contract")
	}

	name := args[0]
	id := args[1]
	forexContract := ""
	interbankContract := ""

	if len(args) > 2 {
		forexContract = args[2]
	}

	if len(args) > 3 {
		interbankContract = args[3]
	}

	bank := bank{Name: name, ID: id, ForexContract: forexContract, InterbankContract: interbankContract}

	bankBytes, _ := json.Marshal(bank)
	err := stub.PutState("bank", bankBytes)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//Invoke is called when external applications invoke the smart contract
func (s *BankChaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "createAccount" {
		return s.createAccount(stub, args)
	} else if function == "queryAccount" {
		return s.queryAccount(stub, args)
	} else if function == "transfer" {
		return s.transfer(stub, args)
	} else if function == "deposit" {
		return s.deposit(stub, args)
	} else if function == "getTransactionHistory" {
		return s.getTransactionHistory(stub, args)
	}

	return shim.Error("Invalid function")
}
