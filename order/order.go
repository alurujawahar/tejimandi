package order

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"

	h "github.com/alurujawahar/tejimandi/httpRequest"
	SmartApi "github.com/angel-one/smartapigo"
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
	ABClient := SmartApi.New(params.ClientCode, params.Password, params.APIKey)
	fmt.Println("Client :- ", ABClient)

	newTotp, err := totp.GenerateCode(params.TOTPKEY, time.Now())
	if err != nil {
		fmt.Println("Failed to generate Totp %v", err)
	}
	session, err := ABClient.GenerateSession(newTotp)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	session.UserProfile, err = ABClient.GetUserProfile()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return ABClient, params, session
}

func OrderBook(A *SmartApi.Client, auth h.ClientParams, session SmartApi.UserSession) {
	url := "https://apiconnect.angelbroking.com/rest/secure/angelbroking/order/v1/getTradeBook"
	method := "GET"
	var payload *strings.Reader
	body := h.HttpRequest(url, method, payload, auth, session)
	fmt.Println("Orders: ", string(body))
}


func PlaceBulkOrder(w http.ResponseWriter, r *http.Request)  {
	exchange := "NSE"
	stocksFilePath := "/Users/alurujawahar/Desktop/angel/tejimandi/stocks_partial.json"
	filepath := "/Users/alurujawahar/Desktop/angel/tejimandi/keys.json"
	ABClient, _, _ := authenticate(filepath)
	var OrderParams []SmartApi.OrderParams
	var ltpParams SmartApi.LTPParams
	res, err := os.Open(stocksFilePath)
	if err != nil {
		fmt.Println(err)
	}

	content, err := io.ReadAll(res)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(content, &OrderParams)
	if err != nil {
		fmt.Println("Unmarshal Failed:", err)
	}
	for _, stk := range OrderParams {
		ltpParams.Exchange = exchange
		ltpParams.SymbolToken = stk.SymbolToken
		ltpParams.TradingSymbol = stk.TradingSymbol
		ltpResp, err := ABClient.GetLTP(ltpParams)
		if err != nil {
			fmt.Println(err)
		}
		stk.Price = ltpResp.Ltp
		stk.SquareOff = fmt.Sprintf("%.2f",ltpResp.Ltp * 1.10)
		stk.StopLoss = fmt.Sprintf("%.2f",ltpResp.Ltp * 0.994)

		fmt.Println("Symbol", stk.TradingSymbol)
		fmt.Println("Squareoff", stk.SquareOff)
		fmt.Println("Stoploss", stk.StopLoss)
		fmt.Println("Price:", stk.Price)
		fmt.Println("")

		if false {
			fmt.Println("Placing Order for Stock: ", stk.TradingSymbol)
			order, err := ABClient.PlaceOrder(stk)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Placed Order ID and Script :- ", order)
		}
	}
}