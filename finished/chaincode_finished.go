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
	stub.PutState("noOfLoansCreated", []byte("0"))
	stub.PutState("loansCreated", []byte("[]"))
	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "create_loan" {	
		return t.createLoan(stub, args)
	} else if function == "update_loanAmount" {
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
		}	
		return t.update_loanAmount(stub, args[0], args[1])
	}else if function == "update_borrower" {
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
		}	
		return t.update_borrower(stub, args[0], args[1])
	}else if function == "update_leadArranger" {
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
		}	
		return t.update_leadArranger(stub, args[0], args[1])
	}else if function == "update_participatingBank" {
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
		}	
		return t.update_participatingBank(stub, args[0], args[1])
	}else if function == "update_status" {
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
		}	
		return t.update_status(stub, args[0], args[1])
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]string, error) {
	fmt.Println("query is running " + function)
	
	// Handle different functions
	if function == "get_loan_details" { //read a variable
		return t.get_loan_details(stub, args)
	}else if function == "get_noOfLoansCreated" { //read a variable
		return t.get_noOfLoansCreated(stub, args)
	}else if function == "get_loansCreated" { //read a variable
		return t.get_loansCreated(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// write - invoke function to write key/value pair
//args:- borrower name
func (t *SimpleChaincode) createLoan(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	bytes, _ := stub.GetState("noOfLoansCreated")
	noOfLoansCreated_ := string(bytes)
	fmt.Println("noOfLoansCreated")
	fmt.Println(noOfLoansCreated_)
	var err error
	noOfLoansCreated, err := strconv.Atoi(noOfLoansCreated_)
	if err != nil{ return nil, errors.New("Invalid value for noOfLoansCreated") }
	noOfLoansCreated += 1
	
	var key string
	
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}
	
	loanID := "loanID_" + noOfLoansCreated_
	loanID_             := "\"loanID\":\""+loanID+"\", "							// Variables to define the JSON
	loanAmount         := "\"loanAmount\":\"UNDEFINED\", "
	disbursedAmoun     := "\"disbursedAmoun\":\"UNDEFINED\", "
	repayedAmount      := "\"repayedAmount\":\"UNDEFINED\", "
	borrower_          := "\"borrower\":\""+args[0]+"\", "
	leadArranger   	   := "\"leadArranger\":\"UNDEFINED\", "
	participatingBank  := "\"participatingBank\":\"UNDEFINED\", "
	Status             := "\"Status\":0, "
        	
	loan_json := "{"+loanID_+loanAmount+disbursedAmoun+repayedAmount+borrower_+leadArranger+participatingBank+Status+"}" 	// Concatenates the variables to create the total JSON object
	fmt.Println("loan_json")
	fmt.Println(loan_json)
	
	fmt.Println("running write()")

	key = loanID //rename for funsies
	
	value, err_ := json.Marshal(loan_json)
	if err_ != nil { fmt.Printf("CREATE_LOAN: Error marshalling loanJson: %s", err_); return nil, errors.New("Error marshalling loanJson")}
	
	err = stub.PutState(key, value) //write the new loan json byte array into the chaincode state
	if err != nil {
		fmt.Println("Unable to put in state the new loan")
		return nil, err
	}
	
	loan_arr_str, err_ := stub.GetState("loansCreated")
	if err_ != nil {
		fmt.Println("Unable to get state loansCreated")
		return nil, err
	}
	var loan_array []string
	json.Unmarshal(loan_arr_str, &loan_array)
	loan_array = append(loan_array, loanID)
	fmt.Println("loan_array")
	fmt.Println(loan_array)
	bytes_la, err_la := json.Marshal(loan_array)
	if err_la != nil { fmt.Printf("CREATE_LOAN: Error marshalling loan_array: %s", err); return nil, errors.New("Error marshalling loan_array")}
	
	stub.PutState("loansCreated", bytes_la)
	
	noOfLoansCreatedBytes := []byte(string(noOfLoansCreated))
	err = stub.PutState("noOfLoansCreated", noOfLoansCreatedBytes)
	if err == nil {fmt.Println("Successfully written to noOfLoansCreated")}
	
	return bytes_la, nil
}

// read - query function to read key/value pair
//args:- loanID
func (t *SimpleChaincode) get_loan_details(stub shim.ChaincodeStubInterface, args []string) (loan, error) {
	
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
	
	var valJson loan
	json.Unmarshal(valAsbytes, &valJson)
	fmt.Println("retrieved loan json")
	fmt.Println(valJson)
	
	//return valAsbytes, nil
	return valJson, nil
}

// read - query function to read key/value pair
//args:- none
func (t *SimpleChaincode) get_noOfLoansCreated(stub shim.ChaincodeStubInterface, args []string) (int, error) {
	
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
//args:- none
func (t *SimpleChaincode) get_loansCreated(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	
	var key, jsonResp string
	var err error

	key = "loansCreated"
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	var noOfLoansCreated_ string
	json.Unmarshal(valAsbytes, &noOfLoansCreated_)
	noOfLoansCreated := strconv.Atoi(noOfLoansCreated_)
	fmt.Println("No of loans created")
	fmt.Println(noOfLoansCreated)
	//return valAsbytes, nil
	return noOfLoansCreated, nil
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
	
	loanJson, _ := stub.GetState(loanID)
	var loanjsonVal loan
	json.Unmarshal(loanJson, &loanjsonVal)
	loanjsonVal.Status = new_status
	
	bytes, err := json.Marshal(loanJson)
	if err != nil { fmt.Printf("UPDATE_STATUS: Error marshalling loanJson: %s", err); return nil, errors.New("Error marshalling loanJson")}
	
	err  = stub.PutState(loanID, bytes)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_STATUS: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return bytes, nil

}

//=================================================================================================================================
//	 update_borrower
//=================================================================================================================================
func (t *SimpleChaincode) update_borrower(stub shim.ChaincodeStubInterface,loanID string, borrower string) ([]byte, error) {
        var err error
		
	loanJson, _ := stub.GetState(loanID)
	var loanjsonVal loan
	json.Unmarshal(loanJson, &loanjsonVal)
	loanjsonVal.borrower = borrower
	
	bytes, err := json.Marshal(loanJson)
	if err != nil { fmt.Printf("UPDATE_BORROWER: Error marshalling loanJson: %s", err); return nil, errors.New("Error marshalling loanJson")}
	
	err  = stub.PutState(loanID, bytes)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_BORROWER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return bytes, nil

}

//=================================================================================================================================
//	 update_leadArranger
//=================================================================================================================================
func (t *SimpleChaincode) update_leadArranger(stub shim.ChaincodeStubInterface,loanID string, arranger string) ([]byte, error) {
        var err error
		
	loanJson, _ := stub.GetState(loanID)
	var loanjsonVal loan
	json.Unmarshal(loanJson, &loanjsonVal)
	loanjsonVal.leadArranger = arranger
	
	bytes, err := json.Marshal(loanJson)
	if err != nil { fmt.Printf("UPDATE_PARTICIPATING: Error marshalling loanJson: %s", err); return nil, errors.New("Error marshalling loanJson")}

	err  = stub.PutState(loanID, bytes)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_LEADARRANGER: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return bytes, nil

}

//=================================================================================================================================
//	 update_participatingBank
//=================================================================================================================================
func (t *SimpleChaincode) update_participatingBank(stub shim.ChaincodeStubInterface,loanID string, participatingBank string) ([]byte, error) {
        var err error
		
	loanJson, _ := stub.GetState(loanID)
	var loanjsonVal loan
	json.Unmarshal(loanJson, &loanjsonVal)
	loanjsonVal.participatingBank = participatingBank
	
	bytes, err := json.Marshal(loanJson)
	if err != nil { fmt.Printf("UPDATE_PARTICIPATING: Error marshalling loanJson: %s", err); return nil, errors.New("Error marshalling loanJson")}
	
	err  = stub.PutState(loanID, bytes)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_PARTICIPATING: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return bytes, nil

}

//=================================================================================================================================
//	 update_loanAmount
//=================================================================================================================================
func (t *SimpleChaincode) update_loanAmount(stub shim.ChaincodeStubInterface,loanID string, amount string) ([]byte, error) {
        var err error
	
	new_amount, err := strconv.Atoi(string(amount))
	
	loanJson, _ := stub.GetState(loanID)
	var loanjsonVal loan
	json.Unmarshal(loanJson, &loanjsonVal)
	loanjsonVal.loanAmount = new_amount
	
	bytes, err := json.Marshal(loanjsonVal)
	if err != nil { fmt.Printf("UPDATE_LOANAMOUNT: Error marshalling loanJson: %s", err); return nil, errors.New("Error marshalling loanJson")}
	
	err  = stub.PutState(loanID, bytes)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_LOANAMOUNT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return bytes, nil

}


