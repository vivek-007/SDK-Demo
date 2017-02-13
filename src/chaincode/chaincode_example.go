package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Owner struct stores owner specific details
type Owner struct {
	Aadhar    int64   `json:"aadhar"`
	SurveyNos []int64 `json:"surveyNumbers"`
}

// Survey struct stores property specific details
type Survey struct {
	Area     int64    `json:"area"`
	Location string   `json:"location"`
	Owners   []string `json:"owners"`
}

// Init : Adds initial block to chaincode on blockchain network
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval))) //making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Run : Entry point for all the Invoke functions
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// Invoke : Adds a new block to the blockchain network
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function == "init" {
		// Initial block - Puts 'abc' and '99'
		return t.Init(stub, "init", args)
	} else if function == "initProperty" {
		// Pushes property details to Blockchain network
		return t.initProperty(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error

	return nil, errors.New("Received unknown function invocation")
}

// initProperty : Registers a new property
func (t *SimpleChaincode) initProperty(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expected 5 arguments")
	}

	fmt.Println("- start init property")

	var owner Owner
	var survey Survey
	var err error

	// Setting keys
	var ownerName = args[0]
	var surveyNumber int64
	surveyNumber, _ = strconv.ParseInt(args[2], 10, 64)

	// Get owner's state from blockchain network
	valAsBytes, _ := stub.GetState(args[0])

	// Unmarshalling the owner's state
	var retrieveSurveyNos Owner
	json.Unmarshal(valAsBytes, &retrieveSurveyNos)

	// Setting the owner object
	if retrieveSurveyNos.Aadhar == 0 {
		owner.Aadhar, _ = strconv.ParseInt(args[1], 10, 64)
		owner.SurveyNos = append(owner.SurveyNos, surveyNumber)
	} else {
		owner.Aadhar = retrieveSurveyNos.Aadhar
		owner.SurveyNos = append(retrieveSurveyNos.SurveyNos, surveyNumber)
	}

	// Marshalling the owner object
	bytes, _ := json.Marshal(owner)
	err = stub.PutState(ownerName, bytes)
	if err != nil {
		return nil, errors.New("Putstate failed")
	}

	// Get the survey state from blockchain network
	surveyAsBytes, _ := stub.GetState(args[2])

	// Unmarshalling the survey object
	var retrieveSurvey Survey
	json.Unmarshal(surveyAsBytes, &retrieveSurvey)

	// Setting the survey object
	if retrieveSurvey.Area == 0 {
		survey.Location = args[3]
		survey.Area, _ = strconv.ParseInt(args[4], 10, 64)
		survey.Owners = append(survey.Owners, ownerName)
	} else {
		survey.Location = retrieveSurvey.Location
		survey.Area = retrieveSurvey.Area
		survey.Owners = append(retrieveSurvey.Owners, ownerName)
	}

	// Marshalling the survey object
	bytesSurvey, _ := json.Marshal(survey)
	err = stub.PutState(args[2], bytesSurvey)
	if err != nil {
		return nil, errors.New("Putstate failed")
	}

	fmt.Println("- end init property")
	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "readInit" { // read init (key: 'abc') value, used for tracking pre-flight check
		return t.readInit(stub, args)
	} else if function == "readOwner" { // read a owner's details
		return t.readOwner(stub, args)
	} else if function == "readSurvey" { // read survey details
		return t.readSurvey(stub, args)
	}
	fmt.Println("query did not find func: " + function) //error

	return nil, errors.New("Received unknown function query - Team PSL")
}

// readInit - used for reading init value, i.e 'abc' and '99'
func (t *SimpleChaincode) readInit(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments")
	}

	valAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Couldn't find init value, Please pass correct key")
	}

	return valAsBytes, nil
}

// read a owner's details
func (t *SimpleChaincode) readOwner(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Invalid numbers of arguments. Expected ")
	}

	var owner Owner

	// Get owner's details from chaincode state
	valAsBytes, _ := stub.GetState(args[0])
	// Unmarshalling the owner's details
	json.Unmarshal(valAsBytes, &owner)

	if owner.Aadhar == 0 {
		return nil, errors.New("Owner doesn't exist")
	}

	//Marshalling the owner object
	bytes, _ := json.Marshal(owner)
	a := []byte("'")
	str := append(a, bytes...)
	str = append(str, a...)
	return []byte(str), nil // returns JSON of owner's details
}

// read a survey details
func (t *SimpleChaincode) readSurvey(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Invalid numbers of arguments. Expected ")
	}

	var survey Survey
	// Get survey details from chaincode state
	valAsBytes, _ := stub.GetState(args[0])
	// Unmarshalling the owner's details
	json.Unmarshal(valAsBytes, &survey)

	if survey.Area == 0 {
		return nil, errors.New("Survey number doesn't exist")
	}

	//Marshalling the survey object
	bytes, _ := json.Marshal(survey)
	a := []byte("'")
	str := append(a, bytes...)
	str = append(str, a...)
	return []byte(str), nil // returns JSON of survey details
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
