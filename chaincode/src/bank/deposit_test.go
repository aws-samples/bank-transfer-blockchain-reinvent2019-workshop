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
	"github.com/google/uuid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPay(t *testing.T) {

	bankStub := shim.NewMockStub("bank", new(BankChaincode))

	uid := uuid.New().String()

	response := bankStub.MockInit(uid, [][]byte{[]byte("CloudBank"), []byte("0001"), []byte("forex")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	uid = uuid.New().String()
	response = bankStub.MockInvoke(uid, [][]byte{[]byte("createAccount"),
		[]byte("Bob Jones"), []byte("0001"), []byte("0"), []byte("USD")})

	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//perform transfer
	uid = uuid.New().String()
	response = bankStub.MockInvoke(uid, [][]byte{[]byte("deposit"), []byte("0001"), []byte("500")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//query from account
	uid = uuid.New().String()
	response = bankStub.MockInvoke(uid, [][]byte{[]byte("queryAccount"),
		[]byte("0001")})

	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	acc := &account{}

	err := json.Unmarshal(response.GetPayload(), acc)

	if err != nil {
		panic(err)
	}

	validateBalance, _ := decimal.NewFromString("500")
	assert.Equal(t, validateBalance, acc.Balance, "incorrect balance")
}
