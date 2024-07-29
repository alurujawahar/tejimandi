package order

import (
	// "context"
	"encoding/json"
	"fmt"
	"io"
	// "log"
	"os"
	"strings"
	// "time"

	h "github.com/alurujawahar/tejimandi/httpRequest"
	SmartApi "github.com/angel-one/smartapigo"
)

// type ClientParams struct {
// 	ClientCode  string `json:"client"`
// 	Password  string `json:"password"`
// 	APIKey  string `json:"api_key"`
// 	TOTPKEY string `json:"totp"`
// }

func OrderBook(A *SmartApi.Client, auth h.ClientParams, session SmartApi.UserSession) {
	url := "https://apiconnect.angelbroking.com/rest/secure/angelbroking/order/v1/getTradeBook"
	method := "GET"
	var payload *strings.Reader
	body := h.HttpRequest(url, method, payload, auth, session)
	fmt.Println("Orders: ", string(body))
}


func PlaceBulkOrder(A *SmartApi.Client, s string, exchange string)  {
	var OrderParams []SmartApi.OrderParams
	var ltpParams SmartApi.LTPParams
	res, err := os.Open(s)
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
		ltpResp, err := A.GetLTP(ltpParams)
		if err != nil {
			fmt.Println(err)
		}
		stk.Price = ltpResp.Ltp

		if stk.ProductType == "BO" {
			stk.SquareOff = fmt.Sprintf("%.2f",ltpResp.Ltp * 1.10)
			stk.StopLoss = fmt.Sprintf("%.2f",ltpResp.Ltp * 0.98)
		}

		if true {
			if stk.Executed == false {
				fmt.Println("Placing Order for Stock: ", stk.TradingSymbol)
				order, err := A.PlaceOrder(stk)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("Placed Order ID and Script :- ", order)
				stk.Executed = true
			}
		}
	}
}