package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

type SmartContract struct {
	contractapi.Contract
}
type Book struct {
	ID       string  `json:"ID"`
	BookName string  `json:"bookName"`
	Author   string  `json:"author"`
	Price    float64 `json:"price"`
	Number   int     `json:"number"`
	Owner    string  `json:"owner"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	books := []Book{
		{"1001", "西游记", "吴承恩", 39.9, 30, "library"},
		{"1002", "水浒传", "施耐庵", 49.9, 20, "library"},
		{"1003", "三国演义", "罗贯中", 29.9, 10, "library"},
		{"1004", "红楼梦", "曹雪芹", 45.0, 50, "library"},
		{"1005", "斗罗大陆", "唐家三少", 89.9, 5, "library"},
	}

	for _, book := range books {
		bookJSON, err := json.Marshal(book)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(book.ID, bookJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	return nil
}

func (s *SmartContract) BookExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	bookJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return bookJSON != nil, nil
}
func (s *SmartContract) CreateBook(ctx contractapi.TransactionContextInterface, id string, bookName string, author string, price float64, number int, owner string) error {
	exists, err := s.BookExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the book %s already exists", id)
	}

	book := Book{
		ID:       id,
		BookName: bookName,
		Author:   author,
		Price:    price,
		Number:   number,
		Owner:    owner,
	}
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, bookJSON)
}

func (s *SmartContract) QueryBook(ctx contractapi.TransactionContextInterface, id string) (*Book, error) {
	bookJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if bookJSON == nil {
		return nil, fmt.Errorf("the book %s does not exist", id)
	}

	var book Book
	err = json.Unmarshal(bookJSON, &book)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (s *SmartContract) UpdateBook(ctx contractapi.TransactionContextInterface, id string, bookName string, author string, price float64, number int, owner string) error {
	exists, err := s.BookExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the book %s does not exist", id)
	}

	// overwriting original asset with new asset
	book := Book{
		ID:       id,
		BookName: bookName,
		Author:   author,
		Owner:    owner,
		Price:    price,
		Number:   number,
	}
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, bookJSON)
}

func (s *SmartContract) DeleteBook(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.BookExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the book %s does not exist", id)
	}
	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) TransferBook(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	book, err := s.QueryBook(ctx, id)
	if err != nil {
		return err
	}

	book.Owner = newOwner
	bookJSON, err := json.Marshal(book)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, bookJSON)
}

func (s *SmartContract) queryAllBooks(ctx contractapi.TransactionContextInterface) ([]*Book, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var books []*Book
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var book Book
		err = json.Unmarshal(queryResponse.Value, &book)
		if err != nil {
			return nil, err
		}
		books = append(books, &book)
	}

	return books, nil
}
func main() {
	bookChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := bookChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
