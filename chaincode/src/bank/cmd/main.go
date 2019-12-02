package main

import (
	"bank"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func main() {
	err := shim.Start(new(bank.BankChaincode))
	if err != nil {
		fmt.Printf("Error creating new BankChaincodet: %s", err)
	}

}
