package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
)

//Args struct to passed to server
type Args struct {
	Budget          float64
	StockpercentMap map[string]int
}

//PortfolioResponsedata to be send from server to client
type PortfolioResponsedata struct {
	//E.g. “GOOG:100:+$520.25”, “YHOO:200:-$30.40”
	Stocksbought       string
	CurrentMarketValue float64
	UnvestedAmount     float64
}

//Buyresponse to be send to client
type Buyresponse struct {
	TradeID        int
	Stocksbought   string
	UnvestedAmount float64
}

//input2 for passing trading id
var input2 int
var c *rpc.Client
var err error
var client net.Conn

func main() {
	//	establishConnection()
	client, _ := net.Dial("tcp", "127.0.0.1:1234")
	c := jsonrpc.NewClient(client)
	fmt.Println("Welcome to the Yahoo Finance API .What would you like to do ?")
	for {
		var option int
		fmt.Println("SELECT:")
		fmt.Println("1. Buy the stocks from market.")
		fmt.Println("2. Check the Portfolio Loss/Gain")

		fmt.Scan(&option)
		fmt.Println("=============================================")
		switch option {
		/*case '1':
		os.Exit(0)*/
		case 1:
			optionfirst(client, c)
		case 2:
			optionsecond(client, c)
		default:
			fmt.Println("You have entered wrong option")
		}
	}
}

func optionfirst(client net.Conn, c *rpc.Client) {
	var stockipstr string
	var Budget float64
	fmt.Println("Enter the stock symbol and the percentage of your budget you wish to allocate with them")
	fmt.Scan(&stockipstr)
	fmt.Println("Enter your budget for this trasaction")
	fmt.Scan(&Budget)

	sStocknum := strings.Split(stockipstr, ",")
	count := 0
	//BuyallocationMap consists of stocks n % for eahc stock
	StockpercentMap := make(map[string]int)
	for _, v := range sStocknum {
		sSplited := strings.Split(v, ":")
		sSplitnumper := strings.Split(sSplited[1], "%")
		i, err := strconv.Atoi(sSplitnumper[0])
		if err != nil {
			// handle error
			fmt.Println(err)
			os.Exit(2)
		}
		StockpercentMap[sSplited[0]] = i

		count = count + i
	}
	if count != 100 {
		fmt.Println("Sum of Stock Percentages should be 100")
		os.Exit(2)
	}
	args := &Args{Budget, StockpercentMap}
	var reply Buyresponse

	if err != nil {
		log.Fatal("dialing:", err)
	}

	err = c.Call("StockMarket.BuyingStocks", args, &reply)

	if err != nil {
		log.Fatal("Error While Buying Stocks:", err)
	}
	fmt.Println("The summary of your stocks purchase is:", reply)
	fmt.Println("The Trade id for transaction is:", reply.TradeID)
	fmt.Println("Stocks details are: ", reply.Stocksbought)
	fmt.Print("The Unvested Amount: ")
	fmt.Printf("%.2f", reply.UnvestedAmount)
	fmt.Println()
}

func optionsecond(client net.Conn, c *rpc.Client) {
	var tradeinput int
	//c := jsonrpc.NewClient(client)
	fmt.Println("Enter the trade ID for which you wish to see the portfolio ")
	fmt.Scan(&tradeinput)
	var viewportfolioResponse PortfolioResponsedata
	err = c.Call("StockMarket.ViewStockPortfolio", &tradeinput, &viewportfolioResponse)
	if err != nil {
		log.Fatal("Error While ViewStockPortfolio:", err)
	}
	fmt.Println("The netsummary of the Viewportfolio for loss/gain is :", viewportfolioResponse)
	fmt.Println("Stock data: ", viewportfolioResponse.Stocksbought)
	fmt.Println("Current market price: ", viewportfolioResponse.CurrentMarketValue)
	fmt.Print("The Unvested Amount: ")
	fmt.Printf("%.2f", viewportfolioResponse.UnvestedAmount)
	fmt.Println()
}
