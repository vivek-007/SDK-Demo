package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// For keeping track of all users and survey numbers
var ownerIndexStr = "_ownerIndex"
var surveyIndexStr = "_surveyIndex"

// Owner struct stores owner specific details
type Owner struct {
	Name      string  `json:"name"`
	Aadhar    int64   `json:"aadhar"`
	SurveyNos []int64 `json:"surveyNumbers"`
}

// Survey struct stores property specific details
type Survey struct {
	SurveyNo int64    `json:"surveyNo"`
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

	// Initialize owner and survey index
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty) //marshal an emtpy array of strings to clear the index
	err = stub.PutState(ownerIndexStr, jsonAsBytes)
	if err != nil {
		return nil, errors.New("Error initializing owner index")
	}

	var emptySurvey []int64
	jsonAsBytes, _ = json.Marshal(emptySurvey) //marshal an emtpy array of int64 to clear the index
	err = stub.PutState(surveyIndexStr, jsonAsBytes)
	if err != nil {
		return nil, errors.New("Error initializing survey index")
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
	} else if function == "transfer" {
		// Transfers property from one owner to another
		return t.transfer(stub, args)
	} else if function == "debug" {
		// only fot testing transfer function
		return t.debug(stub, args)
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
	owner.Name = ownerName
	var surveyNumber int64
	surveyNumber, _ = strconv.ParseInt(args[2], 10, 64)
	survey.SurveyNo = surveyNumber

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
		return nil, errors.New("Property already exists")
	}

	// Marshalling the owner object
	bytes, _ := json.Marshal(owner)
	err = stub.PutState(ownerName, bytes)
	if err != nil {
		return nil, errors.New("Putstate failed")
	}

	// Adding owner to ownerIndex
	ownerIndexAsBytes, _ := stub.GetState(ownerIndexStr)
	var ownerIndex []string
	json.Unmarshal(ownerIndexAsBytes, &ownerIndex)
	flag := true
	if len(ownerIndex) == 0 {
		ownerIndex = append(ownerIndex, ownerName)
		flag = false
	} else {
		for i := 0; i < len(ownerIndex); i++ {
			if ownerIndex[i] == ownerName {
				flag = false
			}
		}
	}
	if flag {
		ownerIndex = append(ownerIndex, ownerName)
	}
	bytesOwnerIndex, _ := json.Marshal(ownerIndex)
	err = stub.PutState(ownerIndexStr, bytesOwnerIndex)
	if err != nil {
		return nil, errors.New("Putstate failed")
	}

	// Marshalling the survey object
	bytesSurvey, _ := json.Marshal(survey)
	err = stub.PutState(args[2], bytesSurvey)
	if err != nil {
		return nil, errors.New("Putstate failed")
	}

	// Adding survey to surveyIndex
	surveyIndexAsBytes, _ := stub.GetState(surveyIndexStr)
	var surveyIndex []int64
	json.Unmarshal(surveyIndexAsBytes, &surveyIndex)
	surveyIndex = append(surveyIndex, surveyNumber)
	bytesSurveyIndex, _ := json.Marshal(surveyIndex)
	err = stub.PutState(surveyIndexStr, bytesSurveyIndex)
	if err != nil {
		return nil, errors.New("Putstate failed")
	}

	fmt.Println("- end init property")
	return nil, nil
}

// SliceIndex : find the index of an element in an array (generic)
func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// Debug : testing trasnfer function operations individually
func (t *SimpleChaincode) debug(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	transferSurvey, _ := strconv.ParseInt(args[1], 10, 64)

	// 1. Remove survey number from seller (owner struct)
	sellerAsBytes, _ := stub.GetState(args[0])
	var sellerObj Owner
	var newSellerObj Owner
	json.Unmarshal(sellerAsBytes, &sellerObj)

	newSellerObj.Name = sellerObj.Name
	newSellerObj.Aadhar = sellerObj.Aadhar

	index := SliceIndex(len(sellerObj.SurveyNos), func(i int) bool { return sellerObj.SurveyNos[i] == transferSurvey })
	newSellerObj.SurveyNos = append(sellerObj.SurveyNos[:index], sellerObj.SurveyNos[index+1:]...)

	sellerBytes, _ := json.Marshal(newSellerObj)
	_ = ioutil.WriteFile("output.json", sellerBytes, 0644)
	_ = stub.PutState(args[0], sellerBytes)

	// 2. Add survey number to buyer's survey number array
	buyerAsBytes, _ := stub.GetState(args[2])
	var buyerObj Owner
	var newBuyerObj Owner
	json.Unmarshal(buyerAsBytes, &buyerObj)

	newBuyerObj.Name = buyerObj.Name
	newBuyerObj.Aadhar = buyerObj.Aadhar

	newBuyerObj.SurveyNos = append(buyerObj.SurveyNos, transferSurvey)

	buyerBytes, _ := json.Marshal(newBuyerObj)
	_ = ioutil.WriteFile("output1.json", buyerBytes, 0644)
	_ = stub.PutState(args[2], buyerBytes)

	// 3. Add buyer's name to survey state
	surveyAsBytes, _ := stub.GetState(args[1])
	var surveyObj Survey
	var newSurveyObj Survey
	json.Unmarshal(surveyAsBytes, &surveyObj)

	newSurveyObj.Area = surveyObj.Area
	newSurveyObj.Location = surveyObj.Location
	newSurveyObj.SurveyNo = surveyObj.SurveyNo

	newSurveyObj.Owners = append(surveyObj.Owners, args[2])

	surveyBytes, _ := json.Marshal(newSurveyObj)
	_ = ioutil.WriteFile("output2.json", surveyBytes, 0644)
	_ = stub.PutState(args[1], surveyBytes)
	return nil, nil
}

// Transfer : transfers property from one owner to another
func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expected 3 arguments")
	}

	// Set keys
	sellerName := args[0]
	buyerName := args[2]
	transferSurveyNo, _ := strconv.ParseInt(args[1], 10, 64)

	// 1. Remove survey number from seller (owner struct)

	// Fetch seller and unmarshall the result
	sellerAsBytes, _ := stub.GetState(sellerName)
	var sellerObj Owner
	json.Unmarshal(sellerAsBytes, &sellerObj)

	// Remove survey number from seller's survey number array
	var newSellerObj Owner
	newSellerObj.Name = sellerObj.Name
	newSellerObj.Aadhar = sellerObj.Aadhar
	index := SliceIndex(len(sellerObj.SurveyNos), func(i int) bool { return sellerObj.SurveyNos[i] == transferSurveyNo })
	if index != -1 {
		newSellerObj.SurveyNos = append(sellerObj.SurveyNos[:index], sellerObj.SurveyNos[index+1:]...)
	} else {
		err := errors.New("An error occured")
		return nil, err
	}

	// Put the new state of seller into blockchain
	sellerAsBytes, _ = json.Marshal(newSellerObj)
	_ = stub.PutState(sellerName, sellerAsBytes)

	// 2. Add survey number to buyer's survey number array

	// Fetch buyer and unmarshall the result
	buyerAsBytes, _ := stub.GetState(buyerName)
	var buyerObj Owner
	json.Unmarshal(buyerAsBytes, &buyerObj)

	// Appending survey number to buyer's survey number array
	buyerObj.SurveyNos = append(buyerObj.SurveyNos, transferSurveyNo)

	// Put the new state of buyer into blockchain
	buyerAsBytes, _ = json.Marshal(buyerObj)
	_ = stub.PutState(buyerName, buyerAsBytes)

	// 3. Add buyer's name to survey state

	// Fetch the survey state and unmarshall the result
	surveyAsBytes, _ := stub.GetState(strconv.FormatInt(transferSurveyNo, 10))
	var survey Survey
	json.Unmarshal(surveyAsBytes, &survey)

	// Appending buyer's name to survey struct's owner array
	survey.Owners = append(survey.Owners, buyerName)

	// Put the new state of survey into blockchain
	surveyAsBytes, _ = json.Marshal(survey)
	_ = stub.PutState(strconv.FormatInt(transferSurveyNo, 10), surveyAsBytes)

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
	} else if function == "readOwnerIndex" { // retrieve all owners
		return t.readOwnerIndex(stub, args)
	} else if function == "readSurveyIndex" { // retrieve all survey details
		return t.readSurveyIndex(stub, args)
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

// fetch entire details of owners on blockchain network
func (t *SimpleChaincode) readOwnerIndex(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// Retrieve all owner names as bytes
	ownerListAsBytes, _ := stub.GetState(ownerIndexStr)
	var ownerList []string
	json.Unmarshal(ownerListAsBytes, &ownerList)

	// fetch all owner details
	var owners []Owner
	for i := 0; i < len(ownerList); i++ {
		bytesOwners, _ := stub.GetState(ownerList[i])
		var owner Owner
		json.Unmarshal(bytesOwners, &owner)
		owners = append(owners, owner)
	}

	// Marshalling and appending ''
	bytes, _ := json.Marshal(owners)
	return []byte(bytes), nil // returns JSON of entire owners on network
}

// fetch entire details of survey numbers on blockchain network
func (t *SimpleChaincode) readSurveyIndex(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	// Retrieve all survey numbers as bytes
	surveyListAsBytes, _ := stub.GetState(surveyIndexStr)
	var surveyList []int64
	json.Unmarshal(surveyListAsBytes, &surveyList)

	// fetch all survey details
	var surveys []Survey
	for i := 0; i < len(surveyList); i++ {
		arg := strconv.FormatInt(surveyList[i], 10)
		bytesSurvey, _ := stub.GetState(arg)
		var survey Survey
		json.Unmarshal(bytesSurvey, &survey)
		surveys = append(surveys, survey)
	}

	// Marshalling and appending ''
	bytes, _ := json.Marshal(surveys)
	return []byte(bytes), nil // returns JSON of entire survey numbers on network
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
