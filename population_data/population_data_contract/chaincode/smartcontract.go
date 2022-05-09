package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type PersonData struct {
	PassportNumber     string `json:"PassportNumber"`
	Name               string `json:"Name"`
	LastName           string `json:"LastName"`
	City               string `json:"City"`
	ResidentialAddress string `json:"ResidentialAddress"`
	PhoneNumber        string `json:"PhoneNumber"`
	FamilyStatus       string `json:"FamilyStatus"`
}

// InitLedger adds a base set of data to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	datas := []PersonData{
		{
			PassportNumber:     "1",
			Name:               "Ildar",
			LastName:           "Zinatulin",
			City:               "Moscow",
			ResidentialAddress: "Some street, house 1, apartment 10",
			PhoneNumber:        "+79999999999",
			FamilyStatus:       "Not married",
		},
		{
			PassportNumber:     "2",
			Name:               "Artem",
			LastName:           "Barger",
			City:               "Moscow",
			ResidentialAddress: "Some street, house 1, apartment 11",
			PhoneNumber:        "+79999999998",
			FamilyStatus:       "No data",
		},
	}

	for _, data := range datas {
		dataJSON, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal. %v", err)
		}

		err = ctx.GetStub().PutState(data.PassportNumber, dataJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) PersonDataExists(ctx contractapi.TransactionContextInterface, passportNumber string) (bool, error) {
	dataJSON, err := ctx.GetStub().GetState(passportNumber)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return dataJSON != nil, nil
}

func (s *SmartContract) CreatePersonData(ctx contractapi.TransactionContextInterface, passportNumber string,
	name string, lastName string, city string, residentialAddress string, phoneNumber string, familyStatus string) error {

	exists, err := s.PersonDataExists(ctx, passportNumber)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the person's data with passport number %s already exists", passportNumber)
	}

	data := PersonData{
		PassportNumber:     passportNumber,
		Name:               name,
		LastName:           lastName,
		City:               city,
		ResidentialAddress: residentialAddress,
		PhoneNumber:        phoneNumber,
		FamilyStatus:       familyStatus,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(passportNumber, dataJSON)
}

func (s *SmartContract) ReadPersonData(ctx contractapi.TransactionContextInterface, passportNumber string) (*PersonData, error) {
	dataJSON, err := ctx.GetStub().GetState(passportNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if dataJSON == nil {
		return nil, fmt.Errorf("the person with passport number %s does not exist", passportNumber)
	}

	var data PersonData
	err = json.Unmarshal(dataJSON, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *SmartContract) UpdatePersonData(ctx contractapi.TransactionContextInterface, passportNumber string, name string,
	lastName string, city string, residentialAddress string, phoneNumber string, familyStatus string) error {

	exists, err := s.PersonDataExists(ctx, passportNumber)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the person with passport number %s does not exist", passportNumber)
	}

	data := PersonData{
		PassportNumber:     passportNumber,
		Name:               name,
		LastName:           lastName,
		City:               city,
		ResidentialAddress: residentialAddress,
		PhoneNumber:        phoneNumber,
		FamilyStatus:       familyStatus,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(passportNumber, dataJSON)
}

func (s *SmartContract) PersonDataChanges(ctx contractapi.TransactionContextInterface, passportNumber string) ([]*PersonData, error) {
	exists, err := s.PersonDataExists(ctx, passportNumber)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the person with passport number %s does not exist", passportNumber)
	}

	iterator, err := ctx.GetStub().GetHistoryForKey(passportNumber)
	var data []*PersonData
	for iterator.HasNext() {
		historyValue, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		var value PersonData
		err = json.Unmarshal(historyValue.Value, &value)
		if err != nil {
			return nil, err
		}
		data = append(data, &value)
	}

	return data, nil
}
