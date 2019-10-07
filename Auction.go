package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

//AuctionDetails model
type AuctionDetails struct {
	Base float64
}

//Auctionee model
type Auctionee struct {
	ID        int
	AuctionID int
	Bidders   string
	Winner    int
	MaxBid    float64
}

//Bidder model
type Bidder struct {
	ID          int
	AuctioneeID int
	Name        string
	BidValue    float64
	Base        float64
}

//Winner Model
type Winner struct {
	MaxBid        float64
	WinnerDetails Bidder
}

func main() {
	handleRequests()
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/newAuction", insertNewAuction)
	http.HandleFunc("/newAuctionee", createAuctionee)
	http.HandleFunc("/newBidders", createBidders)
	http.HandleFunc("/registerBidder", registerBidders)
	http.HandleFunc("/placeBid", placeBid)
	http.HandleFunc("/findWinner", findWinner)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "test"
	dbName := "auction_db"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
}

//Basic Master creation start

func insertNewAuction(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		reqBody := json.NewDecoder(r.Body)
		var auctionDet AuctionDetails
		err := reqBody.Decode(&auctionDet)
		defer r.Body.Close()
		baseValue := auctionDet.Base
		result, err := db.Exec("INSERT INTO Auction_Details(base_value) VALUES(?)", baseValue)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(result)
	}
	defer db.Close()
	fmt.Fprintf(w, "New Auction Created!")
}

func createAuctionee(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var auctionee Auctionee
		err := decoder.Decode(&auctionee)
		result, err := db.Exec("INSERT INTO Auctioneer(auction_id, reg_bidders) VALUES(?, '')", auctionee.AuctionID)
		if err != nil {
			panic(err)
		}
		log.Println(result)
	}
	defer db.Close()
	fmt.Fprintf(w, "New Auctioneer Created!")
}

func createBidders(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var bidder Bidder
		err := decoder.Decode(&bidder)
		result, err := db.Exec("INSERT INTO Bidder(name) VALUES(?)", bidder.Name)
		if err != nil {
			panic(err)
		}
		log.Println(result)
	}
	defer db.Close()
	fmt.Fprintf(w, "New Bidder Created!")
}

//Basic Master creation end
//-----------------------------------------------------------------//
//Workflow starts

//Register bidders with the Auctioneer
func registerBidders(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "PUT" {
		auctioneeID := r.URL.Query().Get("auctionee_id")
		baseBidValue := r.URL.Query().Get("base")
		decoder := json.NewDecoder(r.Body)
		var bidder Bidder
		err := decoder.Decode(&bidder)
		if err != nil {
			panic(err)
		}
		if bidder.ID != 0 {
			result, err := db.Exec("UPDATE Bidder SET auctionee_id=?, base_value=? where bidder_id=?", auctioneeID, baseBidValue, bidder.ID)
			if err != nil {
				panic(err)
			}
			log.Println(result)
			res, err := db.Exec("UPDATE Auctioneer SET reg_bidders=CONCAT(reg_bidders, ?) where id=?", strconv.Itoa(bidder.ID)+", ", auctioneeID)
			if err != nil {
				panic(err)
			}
			defer db.Close()
			log.Println(res)
			log.Println("Bidder Registered !")
		} else {
			panic("Error occured in Bidder decode")
		}
	}

	fmt.Fprintf(w, "Bidder Registered !")
}

//Bidders placing a bid with the Auctioneer
func placeBid(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "PUT" {
		AuctioneeID := r.URL.Query().Get("id")
		decoder := json.NewDecoder(r.Body)
		var bidder Bidder
		err := decoder.Decode(&bidder)
		result, err := db.Exec("UPDATE Bidder SET bid_value=? where bidder_id=? and auctionee_id=?", bidder.BidValue, bidder.ID, AuctioneeID)
		if err != nil {
			panic(err)
		}
		log.Println(result)
		log.Println("Bid placed Succesfully !")
	}
	defer db.Close()
	fmt.Fprintf(w, "Bid placed Succesfully !")
}

//Find a winner of a bid round
func findWinner(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	ID := r.URL.Query().Get("id")
	selDB, err := db.Query("select b.bidder_id, b.auctionee_id, b.name, b.base_value, b.bid_value from auction_db.Bidder b LEFT join auction_db.Auctioneer a on b.auctionee_id = a.id  where a.auction_id = ? ORDER by b.bid_value desc limit 1", ID)
	if err != nil {
		panic(err.Error())
	}
	bid := Bidder{}
	for selDB.Next() {
		var id int
		var auctioneeID int
		var name string
		var baseValue float64
		var bidValue float64

		err = selDB.Scan(&id, &auctioneeID, &name, &baseValue, &bidValue)
		if err != nil {
			panic(err.Error())
		}
		bid.ID = id
		bid.Name = name
		bid.BidValue = bidValue
		bid.AuctioneeID = auctioneeID
		bid.Base = baseValue
	}
	res, err := db.Query("UPDATE Auctioneer SET winner=?, max_bid=? where id=?", bid.ID, bid.BidValue, bid.AuctioneeID)
	if err != nil {
		panic(err.Error())
	}
	log.Println("Winner Declared ", res)
	defer db.Close()
	Winner := Winner{}
	Winner.MaxBid = bid.BidValue
	Winner.WinnerDetails = bid
	fmt.Println("Winner is : ", bid.Name)
	fmt.Println("Max Bid Value : ", bid.BidValue)
	json.NewEncoder(w).Encode(&Winner)
}
