/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"encoding/json"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

//==============================================================================================================================
//	loan - Defines the structure for a loan object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON loanAmount -> Struct loanAmount.
//==============================================================================================================================
type loan struct {
	loanAmount            int `json:"loanAmount"`
	disbursedAmount        int `json:"disbursedAmount"`
	repayedAmount         int `json:"repayedAmount"`
	borrower              string `json:"borrower"`
	leadArranger          string `json:"leadArranger"`
	participatingBank     string `json:"participatingBank"`
	Status                int `json:"status"`
	loanID                 string `json:"loanID"`
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	stub.PutState("noOfLoansCreated", []byte(0))
	stub.PutState("loansCreated", []byte([]))
	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.createLoan(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)
	
	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// write - invoke function to write key/value pair
func (t *SimpleChaincode) createLoan(stub shim.ChaincodeStubInterface, loanID string, args []string) ([]byte, error) {
	noOfLoansCreated := stub.GetState("noOfLoansCreated")
	fmt.println("noOfLoansCreated")
	fmt.println(noOfLoansCreated)
	noOfLoansCreated = strconv.Atoi(string(noOfloansCreated))
	noOfLoansCreated += 1
	
	var key, value string
	
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}
	
	loanID = "loanID" + noOfLoansCreated
	loanID_             := "\"loanID\":\""+loanID+"\", "							// Variables to define the JSON
	loanAmount         := "\"loanAmount\":\"UNDEFINED\", "
	disbursedAmoun     := "\"disbursedAmoun\":\"UNDEFINED\", "
	repayedAmount      := "\"repayedAmount\":\"UNDEFINED\", "
	borrower_          := "\"borrower\":\""+args[0]+"\", "
	leadArranger   	   := "\"leadArranger\":\"UNDEFINED\", "
	participatingBank  := "\"participatingBank\":\"UNDEFINED\", "
	Status             := "\"Status\":0, "
        	
	loan_json := "{"+loanID_+loanAmount+disbursedAmoun+repayedAmount+borrower_+leadArranger+participatingBank+Status+"}" 	// Concatenates the variables to create the total JSON object
	var err error
	fmt.Println("running write()")

	key = loanID //rename for funsies
	value = json.Marshal(loan_json)
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	
	loan_array := stub.GetState("loansCreated")
	loan_array = json.Unmarshal(loan_array)
	append(loan_array, loanID)
	stub.PutState("loansCreated", []byte(json.Marshal(loan_array)))
	
	return []byte(json.Marshal(loan_array)), nil
}

// read - query function to read key/value pair
func (t *SimpleChaincode) get_loan_details(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// read - query function to read key/value pair
func (t *SimpleChaincode) get_noOfLoansCreated(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	
	var key, jsonResp string
	var err error

	key = "noOfLoansCreated"
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// read - query function to read key/value pair
func (t *SimpleChaincode) get_noOfLoansCreated(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	
	var key, jsonResp string
	var err error

	key = "loansCreated"
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

//=================================================================================================================================
//	 Update Functions
//=================================================================================================================================
//	 update_status
//=================================================================================================================================
func (t *SimpleChaincode) update_status(stub shim.ChaincodeStubInterface,loanID string, status string) ([]byte, error) {
        var err error
	
	new_status, err := strconv.Atoi(string(status)) // will return an error if the new vin contains non numerical chars

	if err != nil{ return nil, errors.New("Invalid value passed for status") }
	
	loanJson := stub.GetState(loanID)
	loanJson = json.Unmarshal(loanJson)
	loanJson.Status = new_status
	
	_, err  = stub.Putstate(loanID, []byte(json.Marshal(loanJson)))		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_STATUS: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return []byte(json.Marshal(loanJson)), nil

}

//=================================================================================================================================
//	 update_borrower
//=================================================================================================================================
func (t *SimpleChaincode) update_borrower(stub shim.ChaincodeStubInterface,loanID string, borrower string) ([]byte, error) {
        var err error
		
	loanJson := stub.GetState(loanID)
	loanJson = json.Unmarshal(loanJson)
	loanJson.borrower = borrower
	
	_, err  = stub.Putstate(loanID, []byte(json.Marshal(loanJson)))		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_BORROWER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return []byte(json.Marshal(loanJson)), nil

}

//=================================================================================================================================
//	 update_leadArranger
//=================================================================================================================================
func (t *SimpleChaincode) update_leadArranger(stub shim.ChaincodeStubInterface,loanID string, arranger string) ([]byte, error) {
        var err error
		
	loanJson := stub.GetState(loanID)
	loanJson = json.Unmarshal(loanJson)
	loanJson.leadArranger = leadArranger
	
	_, err  = stub.Putstate(loanID, []byte(json.Marshal(loanJson)))		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_LEADARRANGER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return []byte(json.Marshal(loanJson)), nil

}

//=================================================================================================================================
//	 update_participatingBank
//=================================================================================================================================
func (t *SimpleChaincode) update_participatingBank(stub shim.ChaincodeStubInterface,loanID string, arranger string) ([]byte, error) {
        var err error
		
	loanJson := stub.GetState(loanID)
	loanJson = json.Unmarshal(loanJson)
	loanJson.participatingBank = participatingBank
	
	_, err  = stub.Putstate(loanID, []byte(json.Marshal(loanJson)))		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_PARTICIPATING: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return []byte(json.Marshal(loanJson)), nil

}

//=================================================================================================================================
//	 update_loanAmount
//=================================================================================================================================
func (t *SimpleChaincode) update_loanAmount(stub shim.ChaincodeStubInterface,loanID string, amount string) ([]byte, error) {
        var err error
	
	new_amount, err := strconv.Atoi(string(amount))
	
	loanJson := stub.GetState(loanID)
	loanJson = json.Unmarshal(loanJson)
	loanJson.loanAmount = new_amount
	
	_, err  = stub.Putstate(loanID, []byte(json.Marshal(loanJson)))		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_LOANAMOUNT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return []byte(json.Marshal(loanJson)), nil

}


