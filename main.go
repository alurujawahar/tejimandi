package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
	SmartApi "github.com/angel-one/smartapigo"
	"github.com/pquerna/otp/totp"
	token "github.com/alurujawahar/tejimandi/token"
	order "github.com/alurujawahar/tejimandi/order"
	db "github.com/alurujawahar/tejimandi/database"
	market "github.com/alurujawahar/tejimandi/market"
	h "github.com/alurujawahar/tejimandi/httpRequest"
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



func main() {
	stocksFilePath := "/Users/alurujawahar/Desktop/angel/tejimandi/stocks.json"
	filepath := "/Users/alurujawahar/Desktop/angel/tejimandi/keys.json"
	
	client := db.ConnectMongo()

	//Get Authenticated
	ABClient, authParams, session := authenticate(filepath)

	//Place Bulk Order
	if true {
		order.PlaceBulkOrder(ABClient, stocksFilePath, "NSE", client)
	}

	if true {
		market.MonitorOrders(ABClient, authParams, session, client)
	}

	// if true {
	// 	order.OrderBook(ABClient, authParams, session)
	// }
	if false {
		var ListParams []SmartApi.OrderParams
		instrument_list := token.GetInstrumentList()
		res, err := os.Open(stocksFilePath)
		if err != nil {
			fmt.Println(err)
		}
		content, err := io.ReadAll(res)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(content, &ListParams)
		for _, list := range ListParams {
			token := token.TokenLookUp(list.TradingSymbol , instrument_list, "NSE" )
			fmt.Println(list.TradingSymbol, token )
		}
	}
}