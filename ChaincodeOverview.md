# Bank Transfer Workshop Chaincode

There are three smart contracts for this workshop:
* Bank - a simple representation of a bank that contains bank accounts
* Forex - a contract for providing currency conversion
* Interbank - a contract for routing payments between banks

# Bank - BankChaincode
The bank chaincode is comprised of a number of source code files. Bank.go is the main file which contains the Invoke and Init functions as well as structs used throughout the chaincode. The bank must be initialized with a minimum of two parameters, name and an ID. The name is purely a description. The ID is a string which is used to uniquely identify the bank and is used as part of the Interbank contract to route payments between banks. The ID is analogous to a SWIFT Code or a Bank Identifier Code (BIC) code. Optionally, you can include two further parameters: forexChaincode and interbankChaincode. These are the names of the ForexChaincode and InterbankChaincode chaincode installed on the same peer as the BankChaincode that provide foreign currency exchange and interbank transfer functionality. 

The Invoke function provides four functions that can be invoked. These are:
* createAccount - create a new account on the ledger 
* queryAccount - retrieve that account from the ledger
* deposit - add funds to an account
* transfer - transfer funds between accounts (at the same bank or between accounts)

# Forex - ForexChaincode
The forex chaincode is the simplest of the three chaincodes. It maps a currency pair (e.g. CAD:USD) to an exchange rate. It exposes two functions:
* getForexPair - write currency pair to the ledger
* createForexPair - get a pair from the ledger

#Interbank - InterbankChaincode
The interbank transfer chaincode acts as a router between banks. The bank chaincode can be instantiated with a reference to an interbank contract and that bank can call the interbank contract to transfer funds from one of its accounts to another bank. It does this by storing a mapping between bank IDs and bank contracts. When a transfer is initiated, the interbank chaincode looks up the ID of the recieving bank, retrives the contract for the recieving bank and pays money to the account at that bank. If the currency differs, it will invoke a ForexChaincode instance to convert the currency. 

Interbank chaincdoe exposes two functions:
* interbankTransfer - perform a transfer between banks
* registerRoute - map a bank ID to that bank's chaincode 

# Interaction

The BankChaincode is the base chaincode used to interact with the other chaincodes. You can create
a bank without having to create the others. Without the other chaincode you will only be able to
transfer funds within the bank and using the same currency. By adding a ForexChaincode you can (once
    the appropirate forex pairs have been added) perform intrabank transfers across different
currencies. Once you've added an InterbankChaindode and registered an route to another bank, you can
subsequently make interbank transfers. Transfers are initiated using the transer function on the
BankChaincode. If the Bank ID belongs to another bank or the currency symbol differs between
accounts the transfer method will invoke the other contracts - if provided, otherwise it will return
an error. 

