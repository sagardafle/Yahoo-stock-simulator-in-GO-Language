package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
)

var stocktradeid = 0

var tradeMap = make(map[int]Trade)

//Args struct to passed to server
type Args struct {
	Budget          float64
	StockpercentMap map[string]int
}

//Buyresponse to be send to client
type Buyresponse struct {
	TradeID        int
	Stocksbought   string
	UnvestedAmount float64
}

//PortfolioResponsedata struct
type PortfolioResponsedata struct {
	//E.g. “GOOG:100:+$520.25”, “YHOO:200:-$30.40”
	Stocksbought       string
	CurrentMarketValue float64
	UnvestedAmount     float64
}

//Stock cntg stock's name buying price count
type Stock struct {
	name          string
	purchasePrice float64
	boughtcount   int
}

//Trade for a trade_id it will hold trade details
type Trade struct {
	//id of the trade, static counter increasing per trade.
	Tradingid      int
	UnvestedAmount float64
	Stocks         []Stock
}

//MyResponse structure from Yahoo
type MyResponse struct {
	List struct {
		Meta struct {
			Type  string `json:"type"`
			Start int    `json:"start"`
			Count int    `json:"count"`
		} `json:"meta"`
		Resources []struct {
			Resource struct {
				Classname string `json:"classname"`
				Fields    struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					Utctime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}

//StockMarket structure does buying and dispalying of portfolio
type StockMarket struct{}

//BuyingStocks does buying of stocks called by client
func (t *StockMarket) BuyingStocks(args *Args, reply *Buyresponse) error {

	userbudgetMap := calculatestocks(args)
	stockpriceMap := returnStockData(userbudgetMap)
	buystocks(stockpriceMap, userbudgetMap, reply)

	return nil
}

//ViewStockPortfolio to display portfolio loss or gain for users
func (t *StockMarket) ViewStockPortfolio(X *int, viewportfolioResponse *PortfolioResponsedata) error {

	userbudgetMap := make(map[string]float64)

	trade := tradeMap[(*X)]

	for istock := range trade.Stocks {
		userbudgetMap[trade.Stocks[istock].name] = 0.00

	}
	stockpriceMap := returnStockData(userbudgetMap)

	var buffer bytes.Buffer
	currMktPrice := 0.00
	for istock := range trade.Stocks {
		//userbudgetMap[trade.Stocks[istock].name] = 0.00
		if istock > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(trade.Stocks[istock].name)
		buffer.WriteString(":")

		buffer.WriteString(strconv.Itoa(trade.Stocks[istock].boughtcount))
		buffer.WriteString(":")

		currPrice := stockpriceMap[trade.Stocks[istock].name]
		currMktPrice = (currPrice * float64(trade.Stocks[istock].boughtcount)) + currMktPrice
		if currPrice > (trade.Stocks[istock].purchasePrice) {
			buffer.WriteString("+")
		} else if currPrice < (trade.Stocks[istock].purchasePrice) {
			buffer.WriteString("-")
		} else {
			buffer.WriteString("=")
		}
		buffer.WriteString(strconv.FormatFloat(currPrice, 'f', 2, 64))

	}

	viewportfolioResponse.CurrentMarketValue = currMktPrice
	viewportfolioResponse.UnvestedAmount = trade.UnvestedAmount
	viewportfolioResponse.Stocksbought = buffer.String()

	return nil
}

func calculatestocks(args *Args) map[string]float64 {

	userbudgetMap := make(map[string]float64)

	for stock, percent := range args.StockpercentMap {
		userbudgetMap[stock] = (float64(percent) / 100) * args.Budget
	}
	return userbudgetMap
}

func main() {
	stk := new(StockMarket)
	server := rpc.NewServer()
	server.Register(stk)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {

			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}
}

func buystocks(stockpriceMap map[string]float64, userbudgetMap map[string]float64, reply *Buyresponse) {

	unvested := 0.00
	var trade Trade
	var buffer bytes.Buffer
	//to count diff individual stock in trade
	counter := 0
	stockArr := make([]Stock, len(userbudgetMap))
	for stock, capacity := range userbudgetMap {

		price := stockpriceMap[stock]

		if capacity > price {
			bought, _ := math.Modf(capacity / price)
			unvested = unvested + (capacity - (bought * price))

			// name      purchasePrice  currmktprice boughtcount
			stockArr[counter] = Stock{name: stock, purchasePrice: price, boughtcount: int(bought)}
			counter++
			if counter > 1 {
				buffer.WriteString(",")
			}
			buffer.WriteString(stock)
			buffer.WriteString(":")
			buffer.WriteString(strconv.Itoa(int(bought)))
			buffer.WriteString(":$")
			buffer.WriteString(strconv.FormatFloat(price, 'f', 2, 64))

		} else {
			unvested = unvested + capacity
		}

	}
	if counter > 0 {
		stocktradeid++
		trade.Tradingid = stocktradeid
		trade.UnvestedAmount = unvested
		trade.Stocks = stockArr
		tradeMap[stocktradeid] = trade
	}

	reply.TradeID = stocktradeid
	reply.UnvestedAmount = unvested
	reply.Stocksbought = buffer.String()
}

func returnStockData(userbudgetMap map[string]float64) map[string]float64 {
	var s MyResponse
	var stockpriceMap map[string]float64
	var buffer bytes.Buffer
	//left part of url
	buffer.WriteString("http://finance.yahoo.com/webservice/v1/symbols/")
	//adding the stocks reqd
	stockCounter := 0
	for stock := range userbudgetMap {
		//userbudgetMap[stock] = (float64(percent) / 100) * args.Budget
		if stockCounter > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(stock)
		stockCounter++
	}

	buffer.WriteString("/quote?format=json")

	response, err := http.Get(buffer.String())
	if err != nil {
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {

			os.Exit(1)
		}

		json.Unmarshal([]byte(contents), &s)

		stockpriceMap = make(map[string]float64)
		// stockpriceMap where we key is stockname and value is stockprice
		for i := 0; i < s.List.Meta.Count; i++ {
			f, err1 := strconv.ParseFloat(s.List.Resources[i].Resource.Fields.Price, 64)
			stockpriceMap[s.List.Resources[i].Resource.Fields.Symbol] = f
			if err1 != nil {
				os.Exit(1)
			}
		}
	}
	return stockpriceMap
}
