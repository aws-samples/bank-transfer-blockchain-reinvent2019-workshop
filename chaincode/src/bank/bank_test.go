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
	"forex"
	"github.com/google/uuid"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueryCustomer(t *testing.T) {
	stub := shim.NewMockStub("TestStub", new(BankChaincode))
	uid := uuid.New().String()

	writeResponse := stub.MockInvoke(uid, [][]byte{[]byte("createAccount"),
		[]byte("Bob Jones"), []byte("1"), []byte("400"), []byte("USD")})

	assert.EqualValues(t, shim.OK, writeResponse.GetStatus(), "failed to execute invocation")

	readResponse := stub.MockInvoke(uid, [][]byte{[]byte("queryAccount"),
		[]byte("1")})

	assert.EqualValues(t, shim.OK, readResponse.GetStatus(), "failed to execute invocation")

	resonseAccount := &account{}
	json.Unmarshal(readResponse.GetPayload(), resonseAccount)

	assert.Equal(t, "1", resonseAccount.AccNumber, "account number mismatch")

	validateBalance, _ := decimal.NewFromString("400")
	assert.Equal(t, validateBalance, resonseAccount.Balance, "balance mismatch")

}

func TestTransfer(t *testing.T) {

	fx := new(forex.ForexChaincode)
	b1 := new(BankChaincode)

	forexStub := shim.NewMockStub("forex", fx)
	bankStub := shim.NewMockStub("bank", b1)

	uid := uuid.New().String()
	response := forexStub.MockInit(uid, [][]byte{})

	forexStub.MockPeerChaincode("bank", bankStub)
	bankStub.MockPeerChaincode("forex", forexStub)

	//create forexpair
	uid = uuid.New().String()
	response = forexStub.MockInvoke(uid, [][]byte{[]byte("createUpdateForexPair"), []byte("GBP"), []byte("USD"), []byte("1.20")})

	uid = uuid.New().String()

	response = bankStub.MockInit(uid, [][]byte{[]byte("CloudBank"), []byte("0001"), []byte("forex")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), response.Message)

	uid = uuid.New().String()
	response1 := bankStub.MockInvoke(uid, [][]byte{[]byte("createAccount"),
		[]byte("Bob Jones"), []byte("1"), []byte("0"), []byte("USD")})

	assert.EqualValues(t, shim.OK, response1.GetStatus(), "failed to execute invocation")

	uid = uuid.New().String()
	response2 := bankStub.MockInvoke(uid, [][]byte{[]byte("createAccount"),
		[]byte("Jim Smith"), []byte("2"), []byte("100"), []byte("GBP")})

	assert.EqualValues(t, shim.OK, response2.GetStatus(), "failed to execute invocation")

	//perform transfer
	uid = uuid.New().String()
	response3 := bankStub.MockInvoke(uid, [][]byte{[]byte("transfer"), []byte("2"), []byte("0001"), []byte("1"), []byte("10")})
	assert.EqualValues(t, shim.OK, response3.GetStatus(), "failed to execute invocation")

	//query from account
	uid = uuid.New().String()
	response4 := bankStub.MockInvoke(uid, [][]byte{[]byte("queryAccount"),
		[]byte("2")})

	assert.EqualValues(t, shim.OK, response4.GetStatus(), "failed to execute invocation")

	resonseAccount := &account{}
	err := json.Unmarshal(response4.GetPayload(), resonseAccount)
	if err != nil {
		panic(err)
	}

	validateBalance, _ := decimal.NewFromString("90")
	assert.Equal(t, validateBalance, resonseAccount.Balance, "incorrect balance")

	//Query To Account
	uid = uuid.New().String()

	response = bankStub.MockInvoke(uid, [][]byte{[]byte("queryAccount"),
		[]byte("1")})

	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	resonseAccount = &account{}

	err = json.Unmarshal(response.GetPayload(), resonseAccount)
	if err != nil {
		panic(err)
	}

	validateBalance, _ = decimal.NewFromString("12")
	assert.Equal(t, validateBalance, resonseAccount.Balance, "incorrect balance")
}
