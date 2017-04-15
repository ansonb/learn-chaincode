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
	var key, value string
	
	loanID_             := "\"v5cID\":\""+loanID+"\", "							// Variables to define the JSON
	loanAmount         := "\"loanAmount\":\"UNDEFINED\", "
	disbursedAmoun     := "\"disbursedAmoun\":\"UNDEFINED\", "
	repayedAmount      := "\"repayedAmount\":\"UNDEFINED\", "
	borrower_          := "\"borrower\":\""+caller+"\", "
	leadArranger   	   := "\"leadArranger\":\"UNDEFINED\", "
	participatingBank  := "\"participatingBank\":\"UNDEFINED\", "
	Status             := "\"Status\":0, "
        	
	loan_json := "{"+V5c_ID+loanAmount+disbursedAmoun+repayedAmount+borrower_+leadArranger+participatingBank+Status+"}" 	// Concatenates the variables to create the total JSON object
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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
