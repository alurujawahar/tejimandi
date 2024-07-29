package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
	SmartApi "github.com/angel-one/smartapigo"
	"github.com/pquerna/otp/totp"
	// token "github.com/alurujawahar/tejimandi/token"
	// order "github.com/alurujawahar/tejimandi/order"
	// db "github.com/alurujawahar/tejimandi/database"
	// market "github.com/alurujawahar/tejimandi/market"
	h "github.com/alurujawahar/tejimandi/httpRequest"
	backtest "github.com/alurujawahar/tejimandi/backtest"
)


func authenticate(f string) (*SmartApi.Client, h.ClientParams, SmartApi.UserSession) {
	var params h.ClientParams
	file, err := os.Open(f)
	if err != nil {
		fmt.Println("Unable to open File %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(content, &params)
	// Create New Angel Broking Client
	ABClient := SmartApi.New(params.ClientCode, params.Password, params.APIKey)
	fmt.Println("Client :- ", ABClient)

	newTotp, err := totp.GenerateCode(params.TOTPKEY, time.Now())
	if err != nil {
		fmt.Println("Failed to generate Totp %v", err)
	}
	// User Login and Generate User Session
	session, err := ABClient.GenerateSession(newTotp)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Renew User Tokens using refresh token
	// session.UserSessionTokens, err = ABClient.RenewAccessToken(session.RefreshToken)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// fmt.Println("User Session Tokens :- ", session.UserSessionTokens)

	//Get User Profile
	session.UserProfile, err = ABClient.GetUserProfile()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// fmt.Println("User Profile :- ", session.UserProfile)
	// fmt.Println("User Session Object :- ", session)
	return ABClient, params, session
}

func getDatesInYear(year int) []time.Time {
    var dates []time.Time

    // Start from the first day of the year
    currentDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

    for currentDate.Year() == year {
        dates = append(dates, currentDate)
        currentDate = currentDate.AddDate(0, 0, 1)
    }

    return dates
}

func main() {
	stocksFilePath := "/Users/alurujawahar/Desktop/angel/tejimandi/stocks_partial.json"
	filepath := "/Users/alurujawahar/Desktop/angel/tejimandi/keys.json"
	
	// client := db.ConnectMongo()

	//Get Authenticated
	_, authParams, session := authenticate(filepath)
	// if true {
	// 	symbols := []string{"BANKOFBARODA"}
	// 	startDate := "2023-01-01 09:15"
	// 	endDate := "2023-12-31 15:30"
	// 	initialCapital := 100000.0

	// 	for _, symbol := range symbols {
	// 		backtest.BacktestSymbol(symbol, initialCapital, startDate, endDate, authParams.APIKey, session)
	// 	}
	// }
	// hour, _, _ := time.Now().Clock()
	// if (hour >= 9) && (hour <= 15) {
		//Place Bulk Order
		if false {
			// order.PlaceBulkOrder(ABClient, stocksFilePath, "NSE", client)
		}
		if false {
			// market.MonitorOrders(ABClient, authParams, session, client)
		}
		// if true {
		// 	order.OrderBook(ABClient, authParams, session)
		// }
		if true {
			var ListParams []SmartApi.OrderParams
			year := 2024
			// instrument_list := token.GetInstrumentList()
			res, err := os.Open(stocksFilePath)
			if err != nil {
				fmt.Println(err)
			}
			content, err := io.ReadAll(res)
			if err != nil {
				fmt.Println(err)
			}
			json.Unmarshal(content, &ListParams)
			temp := 0.0
			value := 0.0
			dates := getDatesInYear(year)
			for _, list := range ListParams {
				// token := token.TokenLookUp(list.TradingSymbol , instrument_list, "NSE" )
				// fmt.Println(list.TradingSymbol, token )
				for _, date := range dates {
					// fmt.Println(dateWithTime + " 09:15")
					startDate := date.Format("2006-01-02") + " 09:15"
					endDate := date.Format("2006-01-02") + " 15:30"
					interval := "FIVE_MINUTE"
					initialCapital := 10000.0
					temp = backtest.BacktestSymbol(list.SymbolToken, interval ,initialCapital, startDate, endDate, authParams.APIKey, session)
					value = temp + value
					time.Sleep(2*time.Second)
				}
				
			}

			fmt.Println("VALUE:", value)
		}
	// } else {
	// 	fmt.Println("Can't trade since out of market hours")
	// }
}