package backtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	// "time"

	SmartApi "github.com/angel-one/smartapigo"
	// "github.com/sdcoffey/big"
	// "github.com/sdcoffey/techan"
)


const backTestStartDate string = "2024-01-17 11:00"
const backTestEndDate string = "2024-01-19 16:00"


type HistoricalData struct {
    Date   string  `json:"date"`
    Open   float64 `json:"open"`
    High   float64 `json:"high"`
    Low    float64 `json:"low"`
    Close  float64 `json:"close"`
    Volume int64   `json:"volume"`
}

type hist_input struct {
	Exhange     string `json:"exchange"`
	SymbolToken string `json:"symboltoken"`
	Interval    string `json:"interval"`
	FromDate    string `json:"fromdate"`
	ToDate      string `json:"todate"`
}

type hist_output struct {
	Status    bool            `json:"status"`
	Message   string          `json:"message"`
	Errorcode string          `json:"errorcode"`
	Data      [][]interface{}   `json:"data"`
}

type Candle struct {
	Date   interface{}  `json:"date"`
    Open   interface{} `json:"open"`
    High   interface{} `json:"high"`
    Low    interface{} `json:"low"`
    Close  interface{} `json:"close"`
    Volume interface{}   `json:"volume"`
}


func (hi *hist_input) getCandleData(token string, apiKey string) ([]Candle, error) {
	var histOutput hist_output
	url := "https://apiconnect.angelbroking.com/rest/secure/angelbroking/historical/v1/getCandleData"
	method := "POST"
	jsonData, err := json.Marshal(hi)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	payload := bytes.NewReader(jsonData)

	client := &http.Client{}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Add("X-PrivateKey", apiKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-SourceID", "WEB")
	req.Header.Add("X-ClientLocalIP", "CLIENT_LOCAL_IP")
	req.Header.Add("X-ClientPublicIP", "CLIENT_PUBLIC_IP")
	req.Header.Add("X-MACAddress", "MAC_ADDRESS")
	req.Header.Add("X-UserType", "USER")
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if json.Unmarshal(body, &histOutput) != nil {
		fmt.Println(err)
	}
	var olhc Candle
	var olhcarray []Candle
	for _, i := range histOutput.Data {
			olhc.Date = i[0]
			olhc.Open = i[1] 
			olhc.High = i[2]
			olhc.Low = i[3]
			olhc.Close = i[4]
			olhc.Volume = i[5]
			olhcarray = append(olhcarray, olhc)
	}

	// fmt.Println("OLHC:", olhcarray)

	// var hist []HistoricalData

	// if err := json.NewDecoder(res.Body).Decode(&hist); err != nil {
	// 	fmt.Println(err)
    //     return nil, err
    // }
    return olhcarray, nil
}


func BacktestSymbol(symbol string, interval string, initialCapital float64, backTestStartDate string, backTestEndDate string, apiKey string, session SmartApi.UserSession ) float64 {

	histInput := hist_input{
		FromDate:    backTestStartDate,
		ToDate:      backTestEndDate,
		Exhange:     "NSE",
		SymbolToken: symbol,
		Interval:    interval,
	}
    data, err :=  histInput.getCandleData(session.AccessToken, apiKey)
	if err != nil{
		fmt.Println(err)
	}
    capital := initialCapital
    position := 0

	if len(data) == 0 {
		return 0.0
	}
	fmt.Println("date:", backTestStartDate)
	for _, candle := range data {
        // candle := series.Candles[i]
        // price := candle.Open.(float64)
		buyPrice := data[0].Open.(float64)
		// buyPrice := candle.Open.(float64)
		// fmt.Println("Open Price:", price)
        sellLimit := buyPrice * 1.10
        stopLoss := buyPrice * 0.98

		// fmt.Println("sellLimit:", sellLimit)
		// fmt.Println("StopLoss", stopLoss)

        if position == 0 {
            // Buy at market price
            position = int(capital / buyPrice)
            capital -= float64(position) * buyPrice
            fmt.Printf("Bought %d shares of %s at %.2f on %s\n", position, symbol, buyPrice, candle.Date)
        } else if position == 1 {
			continue
		} else {
            // Check sell conditions
            high := candle.High.(float64)
            low := candle.Low.(float64)
			fmt.Println("High:", high)
			fmt.Println("SellLimit:", sellLimit)
			fmt.Println("Low:", low)
			fmt.Println("Stoploss", stopLoss)
			fmt.Println("==============================")
            if high >= sellLimit {
                capital += float64(position) * sellLimit
                fmt.Printf("Sold %d shares of %s at %.2f on %s\n", position, symbol, sellLimit, candle.Date)
                position = 1
				fmt.Println("Capital(High >= sellLimit):", capital)
            } else if low <= stopLoss {
                capital += float64(position) * stopLoss
                fmt.Printf("Sold %d shares of %s at %.2f (stop loss) on %s\n", position, symbol, stopLoss, candle.Date)
                position = 1
				fmt.Println("Capital(low <= stopLoss):", capital)
            }
        }
    }
	lastCandle := data[len(data)-1]
    finalValue := capital + float64(position)*lastCandle.Close.(float64)
    fmt.Printf("Symbol: %s\n", symbol)
    fmt.Printf("Initial Capital: %.2f\n", initialCapital)
    fmt.Printf("Final Value: %.2f\n", finalValue)
    fmt.Printf("Profit/Loss: %.2f\n", finalValue-initialCapital)
    fmt.Println("-----------------------------------------------------")

	return finalValue-initialCapital
}
