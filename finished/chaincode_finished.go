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
	"regexp"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var logger = shim.NewLogger("CLDChaincode")

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const   AUTHORITY          =  "regulator"
const   BORROWER   	   =  "borrower"
const   LEADARRANGER	   =  "leadarranger"
const   PARTICIPATINGBANK  =  "participatingbank"


//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 6 statuses, this is part of the business logic to determine what can
//					be done to the loan at points in it's lifecycle
//==============================================================================================================================
const   STATE_INIT			=  0
const   STATE_LA_ACCEPT  		=  1
const   STATE_INVITE_PARTICIPATING_BANK	=  2
const   STATE_PARTICIPATING_BANK_ACCEPT	=  3
const   STATE_DISBURSED  		=  4
const   STATE_REPAYED			=  5
//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type  SimpleChaincode struct {
}

//==============================================================================================================================
//	Vehicle - Defines the structure for a car object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type loan struct {
	loanAmount            string `json:"loanAmount"`
	disbursedAmoun        int `json:"disbursedAmount"`
	repayedAmount         int `json:"repayedAmount"`
	borrower              string `json:"borrower"`
	leadArranger          string `json:"leadArranger"`
	participatingBank     bool `json:"participatingBank"`
	Status                int `json:"status"`
	V5cID                 string `json:"v5cID"`
}


//==============================================================================================================================
//	V5C Holder - Defines the structure that holds all the v5cIDs for contracts that have been created.
//				Used as an index when querying all loan contracts.
//==============================================================================================================================

type V5C_Holder struct {
	V5Cs 	[]string `json:"v5cs"`
}

//==============================================================================================================================
//	User_and_eCert - Struct for storing the JSON of a user and their ecert
//==============================================================================================================================

type User_and_eCert struct {
	Identity string `json:"identity"`
	eCert string `json:"ecert"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	/*if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	err := stub.PutState("hello_world", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil*/
		//Args
	//				0
	//			peer_address

	var v5cIDs V5C_Holder

	bytes, err := json.Marshal(v5cIDs)

        if err != nil { return nil, errors.New("Error creating V5C_Holder record") }

	err = stub.PutState("v5cIDs", bytes)

	for i:=0; i < len(args); i=i+2 {
		t.add_ecert(stub, args[i], args[i+1])
	}

	return nil, nil
}

//==============================================================================================================================
//	 add_ecert - Adds a new ecert and user pair to the table of ecerts
//==============================================================================================================================

func (t *SimpleChaincode) add_ecert(stub shim.ChaincodeStubInterface, name string, ecert string) ([]byte, error) {


	err := stub.PutState(name, []byte(ecert))

	if err == nil {
		return nil, errors.New("Error storing eCert for user " + name + " identity: " + ecert)
	}

	return nil, nil

}


//==============================================================================================================================
//	 get_caller - Retrieves the username of the user who invoked the chaincode.
//				  Returns the username as a string.
//==============================================================================================================================

func (t *SimpleChaincode) get_username(stub shim.ChaincodeStubInterface) (string, error) {

    username, err := stub.ReadCertAttribute("username");
	if err != nil { return "", errors.New("Couldn't get attribute 'username'. Error: " + err.Error()) }
	return string(username), nil
}

//==============================================================================================================================
//	 check_affiliation - Takes an ecert as a string, decodes it to remove html encoding then parses it and checks the
// 				  		certificates common name. The affiliation is stored as part of the common name.
//==============================================================================================================================

func (t *SimpleChaincode) check_affiliation(stub shim.ChaincodeStubInterface) (string, error) {
    affiliation, err := stub.ReadCertAttribute("role");
	if err != nil { return "", errors.New("Couldn't get attribute 'role'. Error: " + err.Error()) }
	return string(affiliation), nil

}

//==============================================================================================================================
//	 get_caller_data - Calls the get_ecert and check_role functions and returns the ecert and role for the
//					 name passed.
//==============================================================================================================================

func (t *SimpleChaincode) get_caller_data(stub shim.ChaincodeStubInterface) (string, string, error){

	user, err := t.get_username(stub)

    // if err != nil { return "", "", err }

	// ecert, err := t.get_ecert(stub, user);

    // if err != nil { return "", "", err }

	affiliation, err := t.check_affiliation(stub);

    if err != nil { return "", "", err }

	return user, affiliation, nil
}

//==============================================================================================================================
//	 retrieve_v5c - Gets the state of the data at v5cID in the ledger then converts it from the stored
//					JSON into the loan struct for use in the contract. Returns the loan struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) retrieve_v5c(stub shim.ChaincodeStubInterface, v5cID string) (Vehicle, error) {

	var v loan

	bytes, err := stub.GetState(v5cID);

	if err != nil {	fmt.Printf("RETRIEVE_V5C: Failed to invoke vehicle_code: %s", err); return v, errors.New("RETRIEVE_V5C: Error retrieving vehicle with v5cID = " + v5cID) }

	err = json.Unmarshal(bytes, &v);

    if err != nil {	fmt.Printf("RETRIEVE_V5C: Corrupt vehicle record "+string(bytes)+": %s", err); return v, errors.New("RETRIEVE_V5C: Corrupt vehicle record"+string(bytes))	}

	return v, nil
}

//==============================================================================================================================
// save_changes - Writes to the ledger the loan struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub shim.ChaincodeStubInterface, v loan) (bool, error) {

	bytes, err := json.Marshal(v)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting loan record: %s", err); return false, errors.New("Error converting loan record") }

	err = stub.PutState(v.V5cID, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing loan record: %s", err); return false, errors.New("Error storing loan record") }

	return true, nil
}


//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create Vehicle - Creates the initial JSON for the vehcile and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_loan(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string, v5cID string) ([]byte, error) {
	var v Vehicle

	V5cID              := "\"v5cID\":\""+v5cID+"\", "							// Variables to define the JSON
	loanAmount         := "\"loanAmount\":\"UNDEFINED\", "
	disbursedAmoun     := "\"disbursedAmoun\":\"UNDEFINED\", "
	repayedAmount      := "\"repayedAmount\":\"UNDEFINED\", "
	borrower           := "\"borrower\":\""+caller+"\", "
	leadArranger   	   := "\"leadArranger\":\"UNDEFINED\", "
	participatingBank  := "\"participatingBank\":\"UNDEFINED\", "
	Status             := "\"Status\":0, "

	
	loan_json := "{"+v5ID+loanAmount+disbursedAmoun+repayedAmount+borrower+leadArranger+participatingBank+Status+"}" 	// Concatenates the variables to create the total JSON object

	matched, err := regexp.Match("^[A-z][A-z][0-9]{7}", []byte(v5cID))  				// matched = true if the v5cID passed fits format of two letters followed by seven digits

												if err != nil { fmt.Printf("CREATE_LOAN: Invalid v5cID: %s", err); return nil, errors.New("Invalid v5cID") }

	if 				v5cID  == "" 	 ||
					matched == false    {
																		fmt.Printf("CREATE_LOAN: Invalid v5cID provided");
																		return nil, errors.New("Invalid v5cID provided")
	}

	err = json.Unmarshal([]byte(loan_json), &v)							// Convert the JSON defined above into a vehicle object for go

																		if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(v.V5cID) 								// If not an error then a record exists so cant create a new car with this V5cID as it must be unique

																		if record != nil { return nil, errors.New("Loan already exists") }

	/*if 	caller_affiliation != AUTHORITY {							// Only the regulator can create a new v5c

		return nil, errors.New(fmt.Sprintf("Permission Denied. create_loan. %v === %v", caller_affiliation, AUTHORITY))

	}*/

	_, err  = t.save_changes(stub, v)

																		if err != nil { fmt.Printf("CREATE_LOAN: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	bytes, err := stub.GetState("v5cIDs")

																		if err != nil { return nil, errors.New("Unable to get v5cIDs") }

	var v5cIDs V5C_Holder

	err = json.Unmarshal(bytes, &v5cIDs)

																		if err != nil {	return nil, errors.New("Corrupt V5C_Holder record") }

	v5cIDs.V5Cs = append(v5cIDs.V5Cs, v5cID)


	bytes, err = json.Marshal(v5cIDs)

			if err != nil { fmt.Print("Error creating V5C_Holder record") }

	err = stub.PutState("v5cIDs", bytes)

		if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}

//=================================================================================================================================
//	 Update Functions
//=================================================================================================================================
//	 update_status
//=================================================================================================================================
func (t *SimpleChaincode) update_status(stub shim.ChaincodeStubInterface, v loan, caller string, caller_affiliation string, new_status string) ([]byte, error) {
        
	/*Update state only when hese conditions are met
	if (v.State == STATE_INIT && caller == LEADARRANGER) ||
	(v.State == STATE_LA_ACCEPT && caller == LEADARRANGER) ||
	(v.State == STATE_INVITE_PARTICIPATING_BANK && caller == PARTICIPATINGBANK) || 
	(v.State == STATE_PARTICIPATING_BANK_ACCEPT && (caller == PARTICIPATINGBANK || caller == LEADARRANGER)) ||
	(v.State == STATE_DISBURSED && caller == BORROWER) ||
	{*/
        	v.Status = new_status				// Update to the new value
	//}
	_, err  = t.save_changes(stub, v)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_STATUS: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 update_disbursedAmount
//=================================================================================================================================
func (t *SimpleChaincode) update_disbursedAmount(stub shim.ChaincodeStubInterface, v loan, caller string, caller_affiliation string, new_amount int) ([]byte, error) {
        
	/*Update state only when hese conditions are met
	if (v.State == STATE_INIT && caller == LEADARRANGER) ||
	(v.State == STATE_LA_ACCEPT && caller == LEADARRANGER) ||
	(v.State == STATE_INVITE_PARTICIPATING_BANK && caller == PARTICIPATINGBANK) || 
	(v.State == STATE_PARTICIPATING_BANK_ACCEPT && (caller == PARTICIPATINGBANK || caller == LEADARRANGER)) ||
	(v.State == STATE_DISBURSED && caller == BORROWER) ||
	{*/
        	v.repayedAmount = new_amount				// Update to the new value
	//}
	_, err  = t.save_changes(stub, v)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_DISBURSEDAMOUNT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 update_repayedAmount
//=================================================================================================================================
func (t *SimpleChaincode) update_disbursedAmount(stub shim.ChaincodeStubInterface, v loan, caller string, caller_affiliation string, new_amount int) ([]byte, error) {
        
	/*Update state only when hese conditions are met
	if (v.State == STATE_INIT && caller == LEADARRANGER) ||
	(v.State == STATE_LA_ACCEPT && caller == LEADARRANGER) ||
	(v.State == STATE_INVITE_PARTICIPATING_BANK && caller == PARTICIPATINGBANK) || 
	(v.State == STATE_PARTICIPATING_BANK_ACCEPT && (caller == PARTICIPATINGBANK || caller == LEADARRANGER)) ||
	(v.State == STATE_DISBURSED && caller == BORROWER) ||
	{*/
        	v.disbursedAmount = new_amount				// Update to the new value
	//}
	_, err  = t.save_changes(stub, v)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_REPAYEDAMOUNT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 update_participatingBank
//=================================================================================================================================
func (t *SimpleChaincode) update_participatingBank(stub shim.ChaincodeStubInterface, v loan, caller string, caller_affiliation string, new_bank string) ([]byte, error) {
        
	/*Update state only when hese conditions are met
	if (v.State == STATE_INIT && caller == LEADARRANGER) ||
	(v.State == STATE_LA_ACCEPT && caller == LEADARRANGER) ||
	(v.State == STATE_INVITE_PARTICIPATING_BANK && caller == PARTICIPATINGBANK) || 
	(v.State == STATE_PARTICIPATING_BANK_ACCEPT && (caller == PARTICIPATINGBANK || caller == LEADARRANGER)) ||
	(v.State == STATE_DISBURSED && caller == BORROWER) ||
	{*/
        	v.participatingBank = new_bank				// Update to the new value
	//}
	_, err  = t.save_changes(stub, v)		// Save the changes in the blockchain

	if err != nil { fmt.Printf("UPDATE_REPAYEDAMOUNT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	return nil, nil

}

//=================================================================================================================================
//	 Read Functions
//=================================================================================================================================
//	 get_loan_details
//=================================================================================================================================
func (t *SimpleChaincode) get_loan_details(stub shim.ChaincodeStubInterface, v loan, caller string, caller_affiliation string) ([]byte, error) {

	bytes, err := json.Marshal(v)

																if err != nil { return nil, errors.New("GET_VEHICLE_DETAILS: Invalid vehicle object") }

	/*if caller_affiliation	== BORROWER ||
	caller_affiliation	== LEADARRANGER ||
	caller_affiliation	== PARTICIPATINGBANK{*/

		return bytes, nil
	/*} else {
																return nil, errors.New("Permission Denied. get_vehicle_details")
	}*/

}

//=================================================================================================================================
//	 get_loans
//=================================================================================================================================

func (t *SimpleChaincode) get_loans(stub shim.ChaincodeStubInterface, caller string, caller_affiliation string) ([]byte, error) {
	bytes, err := stub.GetState("v5cIDs")

																			if err != nil { return nil, errors.New("Unable to get v5cIDs") }

	var v5cIDs V5C_Holder

	err = json.Unmarshal(bytes, &v5cIDs)

																			if err != nil {	return nil, errors.New("Corrupt V5C_Holder") }

	result := "["

	var temp []byte
	var v loan

	for _, v5c := range v5cIDs.V5Cs {

		v, err = t.retrieve_v5c(stub, v5c)

		if err != nil {return nil, errors.New("Failed to retrieve V5C")}

		temp, err = t.get_vehicle_details(stub, v, caller, caller_affiliation)

		if err == nil {
			result += string(temp) + ","
		}
	}

	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}

	return []byte(result), nil
}

//=================================================================================================================================
//	 check_unique_v5c
//=================================================================================================================================
func (t *SimpleChaincode) check_unique_v5c(stub shim.ChaincodeStubInterface, v5c string, caller string, caller_affiliation string) ([]byte, error) {
	_, err := t.retrieve_v5c(stub, v5c)
	if err == nil {
		return []byte("false"), errors.New("V5C is not unique")
	} else {
		return []byte("true"), nil
	}
}

// Invoke issuer entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
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
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
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
        
	//var newByte = []byte("Test change....")
	return valAsbytes, nil
}

//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {

	caller, caller_affiliation, err := t.get_caller_data(stub)

	if err != nil { return nil, errors.New("Error retrieving caller information")}


	if function == "create_loan" {
        return t.create_loan(stub, caller, caller_affiliation, args[0])
	} else if function == "ping" {
        return t.ping(stub)
    	} else { 																				// If the function is not a create then there must be a car so we need to retrieve the car.
		argPos := 1

		v, err := t.retrieve_v5c(stub, args[argPos])

        if err != nil { fmt.Printf("INVOKE: Error retrieving v5c: %s", err); return nil, errors.New("Error retrieving v5c") 



		} else if function == "update_make"  	    { return t.update_make(stub, v, caller, caller_affiliation, args[0])
		} else if function == "update_model"        { return t.update_model(stub, v, caller, caller_affiliation, args[0])
		} else if function == "update_reg" { return t.update_registration(stub, v, caller, caller_affiliation, args[0])
		} else if function == "update_vin" 			{ return t.update_vin(stub, v, caller, caller_affiliation, args[0])
        } else if function == "update_colour" 		{ return t.update_colour(stub, v, caller, caller_affiliation, args[0])
		} else if function == "scrap_vehicle" 		{ return t.scrap_vehicle(stub, v, caller, caller_affiliation) }

		return nil, errors.New("Function of the name "+ function +" doesn't exist.")

	}
}

