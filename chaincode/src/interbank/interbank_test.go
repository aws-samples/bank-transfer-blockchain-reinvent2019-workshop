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
	"bank"
	"encoding/json"
	"forex"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric/common/util"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransfer(t *testing.T) {

	fx := new(forex.ForexChaincode)
	b1 := new(bank.BankChaincode)
	ibank := new(InterbankChaincode)

	forexStub := shim.NewMockStub("forex", fx)
	bankStub := shim.NewMockStub("bank", b1)
	ibankStub := shim.NewMockStub("ibank", ibank)

	uid := uuid.New().String()
	response := forexStub.MockInit(uid, [][]byte{})

	uid = uuid.New().String()
	response = ibankStub.MockInit(uid, [][]byte{})

	ibankStub.MockPeerChaincode("forex", forexStub)
	ibankStub.MockPeerChaincode("bank", bankStub)

	//create forexpair
	uid = uuid.New().String()
	stringArgs := []string{"createUpdateForexPair", "GBP", "USD", "1.20"}
	args := util.ArrayToChaincodeArgs(stringArgs)
	response = forexStub.MockInvoke(uid, args)
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//create forexpair
	uid = uuid.New().String()
	stringArgs = []string{"createUpdateForexPair", "USD", "GBP", "0.80"}
	args = util.ArrayToChaincodeArgs(stringArgs)
	response = forexStub.MockInvoke(uid, args)
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//create bank
	uid = uuid.New().String()
	response = bankStub.MockInit(uid, [][]byte{[]byte("CloudBank"), []byte("0001"), []byte("forex")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//create account
	uid = uuid.New().String()
	response = bankStub.MockInvoke(uid, [][]byte{[]byte("createAccount"),
		[]byte("Bob Jones"), []byte("1"), []byte("0"), []byte("USD")})

	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	uid = uuid.New().String()
	response = bankStub.MockInvoke(uid, [][]byte{[]byte("queryAccount"), []byte("1")})

	//register route to bank
	uid = uuid.New().String()
	response = ibankStub.MockInvoke(uid, [][]byte{[]byte("registerRoute"), []byte("0001"), []byte("bank"), []byte("forex")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//perform transfer
	uid = uuid.New().String()

	stringArgs = []string{"interbankTransfer", "1", "0001", "100", "GBP"}
	response = ibankStub.MockInvoke(uid, util.ArrayToChaincodeArgs(stringArgs))
	assert.EqualValues(t, shim.OK, response.GetStatus(), response.Message)

	//Query To Account
	uid = uuid.New().String()

	response = bankStub.MockInvoke(uid, [][]byte{[]byte("queryAccount"),
		[]byte("1")})

	assert.EqualValues(t, shim.OK, response.GetStatus(), response.Message)

	resonseAccount := &account{}

	err := json.Unmarshal(response.GetPayload(), resonseAccount)
	if err != nil {
		panic(err)
	}

	validateBalance, _ := decimal.NewFromString("120")
	assert.Equal(t, validateBalance, resonseAccount.Balance, "incorrect balance")
}
