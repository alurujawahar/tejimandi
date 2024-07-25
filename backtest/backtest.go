package backtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	SmartApi "github.com/angel-one/smartapigo"
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
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

	fmt.Println("OLHC:", olhcarray)

	// var hist []HistoricalData

	// if err := json.NewDecoder(res.Body).Decode(&hist); err != nil {
	// 	fmt.Println(err)
    //     return nil, err
    // }
    return olhcarray, nil
}


func BacktestSymbol(symbol string, initialCapital float64, backTestStartDate string, backTestEndDate string, apiKey string, session SmartApi.UserSession ) {

	histInput := hist_input{
		FromDate:    backTestStartDate,
		ToDate:      backTestEndDate,
		Exhange:     "NSE",
		SymbolToken: "4668",
		Interval:    "ONE_HOUR",
	}
    data, err :=  histInput.getCandleData(session.AccessToken, apiKey)
	if err != nil{
		fmt.Println(err)
	}
    series := techan.NewTimeSeries()
	fmt.Println("===================================================================")
	fmt.Println("Data", data)
	fmt.Println("===================================================================")
    for _, record := range data {
        date, _ := time.Parse("2006-01-02", fmt.Sprintf("%.2f", record.Date))
        candle := techan.NewCandle(techan.NewTimePeriod(date, time.Hour*24))
        candle.OpenPrice = big.NewDecimal(record.Open.(float64))
        candle.ClosePrice = big.NewDecimal(record.Close.(float64))
        candle.MaxPrice = big.NewDecimal(record.High.(float64))
        candle.MinPrice = big.NewDecimal(record.Low.(float64))
        candle.Volume = big.NewDecimal(float64(record.Volume.(float64)))
        series.AddCandle(candle)
    }

    // Simulate the bracket order logic
    capital := initialCapital
    position := 0

    for i := 1; i < series.LastIndex(); i++ {
        candle := series.Candles[i]
        price := candle.ClosePrice.Float()
        sellLimit := price * 1.10
        stopLoss := price * 0.99

        if position == 0 {
            // Buy at market price
            position = int(capital / price)
            capital -= float64(position) * price
            fmt.Printf("Bought %d shares of %s at %.2f on %s\n", position, symbol, price, series.Candles[i].Period.Start.Format("2006-01-02"))
        } else {
            // Check sell conditions
            high := series.Candles[i].MaxPrice.Float()
            low := series.Candles[i].MinPrice.Float()

            if high >= sellLimit {
                capital += float64(position) * sellLimit
                fmt.Printf("Sold %d shares of %s at %.2f on %s\n", position, symbol, sellLimit, series.Candles[i].Period.Start.Format("2006-01-02"))
                position = 0
            } else if low <= stopLoss {
                capital += float64(position) * stopLoss
                fmt.Printf("Sold %d shares of %s at %.2f (stop loss) on %s\n", position, symbol, stopLoss, series.Candles[i].Period.Start.Format("2006-01-02"))
                position = 0
            }
        }
    }
	lastCandle := series.LastCandle()
    finalValue := capital + float64(position)*lastCandle.ClosePrice.Float()
    fmt.Printf("Symbol: %s\n", symbol)
    fmt.Printf("Initial Capital: %.2f\n", initialCapital)
    fmt.Printf("Final Value: %.2f\n", finalValue)
    fmt.Printf("Profit/Loss: %.2f\n", finalValue-initialCapital)
    fmt.Println("-----------------------------------------------------")
}
