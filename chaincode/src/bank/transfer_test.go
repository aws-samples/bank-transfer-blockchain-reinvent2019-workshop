package bank

import (
	"encoding/json"
	"forex"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric/common/util"
	"interbank"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterbankTransfer(t *testing.T) {

	fx := new(forex.ForexChaincode)
	b1 := new(BankChaincode)
	b2 := new(BankChaincode)

	ibank := new(interbank.InterbankChaincode)

	forexStub := shim.NewMockStub("forex", fx)
	bankStub := shim.NewMockStub("bank", b1)
	ibankStub := shim.NewMockStub("ibank", ibank)
	bank2stub := shim.NewMockStub("bank2", b2)

	uid := uuid.New().String()
	response := forexStub.MockInit(uid, [][]byte{})

	uid = uuid.New().String()
	response = ibankStub.MockInit(uid, [][]byte{})

	ibankStub.MockPeerChaincode("forex", forexStub)
	ibankStub.MockPeerChaincode("bank", bankStub)
	ibankStub.MockPeerChaincode("bank2", bank2stub)

	bank2stub.MockPeerChaincode("ibank", ibankStub)

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

	//create bank 1
	uid = uuid.New().String()
	response = bankStub.MockInit(uid, [][]byte{[]byte("CloudBank"), []byte("0001"), []byte("forex")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//create bank 2
	uid = uuid.New().String()
	response = bank2stub.MockInit(uid, [][]byte{[]byte("Bank of Internet"), []byte("0002"), []byte("forex"), []byte("ibank")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//create account on bank 1
	uid = uuid.New().String()
	stringArgs = []string{"createAccount", "Bob Jones", "1", "0", "USD"}

	response = bankStub.MockInvoke(uid, util.ArrayToChaincodeArgs(stringArgs))

	//create account on bank 2
	uid = uuid.New().String()
	response = bank2stub.MockInvoke(uid, [][]byte{[]byte("createAccount"),
		[]byte("Joe Blogs"), []byte("1234567"), []byte("1000"), []byte("USD")})

	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	uid = uuid.New().String()
	response = bankStub.MockInvoke(uid, [][]byte{[]byte("queryAccount"), []byte("1")})

	//register route to bank
	uid = uuid.New().String()
	response = ibankStub.MockInvoke(uid, [][]byte{[]byte("registerRoute"), []byte("0001"), []byte("bank"), []byte("forex")})
	assert.EqualValues(t, shim.OK, response.GetStatus(), "failed to execute invocation")

	//perform transfer
	uid = uuid.New().String()

	stringArgs = []string{"transfer", "1234567", "0001", "1", "1000"}
	response = bank2stub.MockInvoke(uid, util.ArrayToChaincodeArgs(stringArgs))
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

	validateBalance, _ := decimal.NewFromString("1000")
	assert.Equal(t, validateBalance, resonseAccount.Balance, "incorrect balance")
}
