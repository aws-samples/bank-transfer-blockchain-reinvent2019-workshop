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

package forex

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

type forex struct {
	Pair string  `json:"pair"`
	Rate float64 `json:"rate"`
}

//ForexChaincode is the struct that all chaincode methods are associated with
type ForexChaincode struct {
}

//Init method is invokved on installation and upgrade
func (s *ForexChaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

//Invoke is called when external applications invoke the smart contract
func (s *ForexChaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "createUpdateForexPair" {
		return s.createUpdateForexPair(stub, args)
	} else if function == "getForexPair" {
		return s.getForexPair(stub, args)
	}

	return shim.Error("Invalid function")
}

func (s *ForexChaincode) createUpdateForexPair(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Expecting 3 arguments, base currency, counter currency , rate")
	}

	baseCurrency := args[0]
	counterCurrency := args[1]
	rate, err := strconv.ParseFloat(args[2], 64)
	pair := baseCurrency + ":" + counterCurrency

	if err != nil {
		return shim.Error("Unable to parse rate from arg[2]")
	}

	forexPair := forex{Pair: pair, Rate: rate}
	asBytes, _ := json.Marshal(forexPair)
	err = stub.PutState(pair, asBytes)

	if err != nil {
		return shim.Error("Unable to commit pair to ledger")

	}

	return shim.Success(nil)
}

func (s *ForexChaincode) getForexPair(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting the base currency and counter currency")
	}

	baseCurrency := args[0]
	counterCurrency := args[1]
	pair := baseCurrency + ":" + counterCurrency

	pairAsBytes, stubError := stub.GetState(pair)

	if stubError != nil {
		return shim.Error(stubError.Error())

	}
	return (shim.Success(pairAsBytes))
}

func main() {
	err := shim.Start(new(ForexChaincode))
	if err != nil {
		panic("Error creating new ForexChaincode: " + err.Error())
	}
}
