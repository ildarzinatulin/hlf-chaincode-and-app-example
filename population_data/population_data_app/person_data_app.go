package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	log.Println("============ application-golang starts ============")

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		log.Fatalf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
	}

	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		log.Fatalf("Failed to get network: %v", err)
	}

	contract := network.GetContract("person_data")

	log.Println("--> Submit Transaction: InitLedger, function creates the initial set of data on the ledger")
	result, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
	}
	log.Println(string(result))

	for true {
		var cmd string
		fmt.Println("Enter a command: create, read, update, history")
		_, err = fmt.Scan(&cmd)
		if err != nil {
			log.Fatalf("Failed to read command: %v", err)
			return
		}

		switch cmd {
		case "create":
			err = create(contract)
		case "read":
			err = read(contract)
		case "update":
			err = update(contract)
		case "history":
			err = getChanges(contract)
		default:
			fmt.Println("Not corrected command")
			continue
		}
		if err != nil {
			return
		}
	}
}

func getChanges(contract *gateway.Contract) error {
	passportNumber, err := readValue("passport number")
	if err != nil {
		return err
	}

	result, err := contract.EvaluateTransaction("PersonDataChanges", passportNumber)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
	}
	log.Println(string(result))
	return nil
}

func read(contract *gateway.Contract) error {
	passportNumber, err := readValue("passport number")
	if err != nil {
		return err
	}

	result, err := contract.EvaluateTransaction("ReadPersonData", passportNumber)
	if err != nil {
		log.Fatalf("Failed to evaluate transaction: %v", err)
		return err
	}
	log.Println(string(result))
	return nil
}

func create(contract *gateway.Contract) error {
	passportNumber, name, lastName, city, residentialAddress, phoneNumber, familyStatus, err := getPersonData()
	if err != nil {
		return err
	}

	result, err := contract.SubmitTransaction("CreatePersonData", passportNumber, name, lastName, city, residentialAddress, phoneNumber, familyStatus)
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
		return err
	}
	log.Println(string(result))

	return nil
}

func update(contract *gateway.Contract) error {
	passportNumber, name, lastName, city, residentialAddress, phoneNumber, familyStatus, err := getPersonData()
	if err != nil {
		return err
	}

	result, err := contract.SubmitTransaction("UpdatePersonData", passportNumber, name, lastName, city, residentialAddress, phoneNumber, familyStatus)
	if err != nil {
		log.Fatalf("Failed to Submit transaction: %v", err)
		return err
	}
	log.Println(string(result))

	return nil
}

func getPersonData() (passportNumber, name, lastName, city, residentialAddress, phoneNumber, familyStatus string, err error) {
	passportNumber, err = readValue("passport number")
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	name, err = readValue("name")
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	lastName, err = readValue("last name")
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	city, err = readValue("city")
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	residentialAddress, err = readValue("residential address")
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	phoneNumber, err = readValue("phone number")
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	familyStatus, err = readValue("family status")
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	return passportNumber, name, lastName, city, residentialAddress, phoneNumber, familyStatus, nil
}

func readValue(fieldName string) (string, error) {
	var value string
	fmt.Printf("Enter the %s:\n", fieldName)
	_, err := fmt.Scan(&value)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", fieldName, err)
		return "", err
	}
	return value, nil
}

func populateWallet(wallet *gateway.Wallet) error {
	log.Println("============ Populating wallet ============")
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	return wallet.Put("appUser", identity)
}
