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

const apiUrl = "localhost:8080"
const authToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjg3NTc4NzV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"

type StockData struct {
	Token  string `json:"token"`
    Symbol string  `json:"symbol"`
}

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

	var Symbol string
	var Quantity string
	var ltpParams SmartApi.LTPParams
	var responseData []StockData

	exchange := "NSE"
	filepath := "/Users/alurujawahar/Desktop/angel/tejimandi/keys.json"
	ABClient, _, _ := authenticate(filepath)
	
	req, err := http.NewRequest("GET", "http://"+apiUrl+"/stocks", nil)
	req.Header.Add("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Errorf("Error fetching stockts", err)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&responseData)

	for _, stk := range responseData {
		parts := strings.Split(stk.Symbol, ":")
		if len(parts) == 2 {
			Symbol = parts[0]
			Quantity = parts[1]
			fmt.Printf("Symbol: %s, Quantity: %s\n", Symbol, Quantity)
		} else {
			fmt.Println("Invalid data format")
		}
		ltpParams.Exchange = exchange
		ltpParams.SymbolToken = stk.Token
		ltpParams.TradingSymbol = Symbol
		ltpResp, err := ABClient.GetLTP(ltpParams)
		if err != nil {
			fmt.Println(err)
		}

		OrderParams := SmartApi.OrderParams{
			Variety: "NORMAL",
			TradingSymbol: Symbol,
			SymbolToken: stk.Token,
			TransactionType: "BUY",
			Exchange: "NSE",
			OrderType: "MARKET",
			ProductType: "BO",
			Duration: "DAY",
			Price: ltpResp.Ltp,
			SquareOff: fmt.Sprintf("%.2f",ltpResp.Ltp * 1.10),
			StopLoss: fmt.Sprintf("%.2f",ltpResp.Ltp * 0.994),
			Quantity: Quantity,
			Executed: false,
			
		}

		fmt.Println(OrderParams)
		if false {
			fmt.Println("Placing Order for Stock: ", stk.Symbol)
			order, err := ABClient.PlaceOrder(OrderParams)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Placed Order ID and Script :- ", order)
		}
	}
}