package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"interbank"
)

func main() {
	err := shim.Start(new(interbank.InterbankChaincode))
	if err != nil {
		fmt.Printf("Error creating new InterbankChaincode: %s", err)
	}

}
