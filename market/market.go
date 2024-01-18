package market

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	db "github.com/alurujawahar/tejimandi/database"
	h "github.com/alurujawahar/tejimandi/httpRequest"
	SmartApi "github.com/angel-one/smartapigo"
	"go.mongodb.org/mongo-driver/mongo"
)

const stoploss = -0.2
// type ClientParams struct {
// 	ClientCode  string `json:"client"`
// 	Password  string `json:"password"`
// 	APIKey  string `json:"api_key"`
// 	TOTPKEY string `json:"totp"`
// }

type change_input struct {
	Mode string `json:"mode"`
	ExchangeTokens exchange `json:"exchangeTokens"`
}

type exchange struct {
	NSE []string `json:"NSE"`
}
type position struct {
	Status bool `json:"status"`
	Message string `json:"message"`
	Errorcode string `json:"errorcode"`
	Data position_data `json:"data"`
}

type position_data struct {
	Fetched  []fetched `json:"fetched"`
}

type fetched struct {
	Exchange string `json:"exchange"`
	TradingSymbol string `json:"tradingSymbol"`
	SymbolToken string `json:"symbolToken"`
	Ltp  float64 `json:"ltp"` 
	Open float64 `json:"open"`
	High float64 `json:"high"`
	Low float64 `json:"low"`
	Close float64 `json:"close"`
	PercentChange float64 `json:"percentChange"`
}

func getValueChange(token string, symbol string, auth h.ClientParams, session SmartApi.UserSession) float64 {
	var changeInput change_input
	var positionData position
	var percentageChange float64
	url := "https://apiconnect.angelbroking.com/rest/secure/angelbroking/market/v1/quote/"
	method := "POST"

	changeInput.Mode = "FULL"
	changeInput.ExchangeTokens.NSE = append(changeInput.ExchangeTokens.NSE, token)
	// instrument_list := getInstrumentList()

	jsonData, err := json.Marshal(changeInput)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		os.Exit(1)
	}
	payload := strings.NewReader(string(jsonData))
	body := h.HttpRequest(url, method, payload, auth, session)

	json.Unmarshal(body, &positionData)
	// symbol := symbolLookUp(token, instrument_list, "NSE")
		for _, f := range positionData.Data.Fetched {
			fmt.Printf("Pecentage change of token %s is %.2f\n", symbol , f.PercentChange)
			percentageChange = f.PercentChange
		}
	
	return percentageChange
}

func calPercentageChange(symbol string, client *mongo.Client, currentPrice float64) float64 {
	data, _ := db.QueryMongo(client, symbol)
	percentageChangeFromDB := ((currentPrice - data.Price)/data.Price)*100
	return percentageChangeFromDB
}

func MonitorOrders(A *SmartApi.Client, auth h.ClientParams, session SmartApi.UserSession, client *mongo.Client) {
	loopvar := 1
	var exitParams SmartApi.OrderParams
	var ltpparams SmartApi.LTPParams
	for loopvar != 0 {
		positions, err := A.GetPositions()
		if err != nil {
			fmt.Println("Error getting your Positions", err)
			os.Exit(1)
		}
		for _, pos := range positions {
			if err != nil {
				fmt.Println(err.Error())
			}
			if pos.ProductType == "INTRADAY" {
				fmt.Println("Net Value ", pos.NetValue)
			}
			fmt.Println(pos.SymbolToken)
			// percentChange := getValueChange(pos.SymbolToken, pos.Tradingsymbol, auth, session)
			ltpparams.Exchange = pos.Exchange
			ltpparams.SymbolToken = pos.SymbolToken
			ltpparams.TradingSymbol = pos.Tradingsymbol
			ltp, err := A.GetLTP(ltpparams)
			if err != nil {
				fmt.Println("unable to get Ltp for:", pos.Tradingsymbol, ltp)
			}
			percentChange := calPercentageChange(pos.Tradingsymbol, client, ltp.Ltp)
			data, objectId := db.QueryMongo(client, pos.Tradingsymbol)

			//Sell stocks if they are less than stoploss
			if percentChange < stoploss && data.Executed {
				exitParams.Exchange = pos.Exchange
				exitParams.Variety = "NORMAL"
				exitParams.TradingSymbol = pos.Tradingsymbol
				exitParams.SymbolToken = pos.SymbolToken
				exitParams.OrderType = "LIMIT" 
				exitParams.ProductType = "INTRADAY"
				exitParams.Duration = "DAY"
				exitParams.SquareOff = "0"
				exitParams.StopLoss = "0"
				exitParams.Quantity = "1"
				exitParams.Price = ltp.Ltp
				exitParams.TransactionType = "SELL"
				exitParams.Executed = false
				if false {
					orderResponse, err := A.PlaceOrder(exitParams)
					if err != nil {
						fmt.Println("Failed to exit position", err)
					}
					fmt.Println("Successfully exited trading Symbol", pos.Tradingsymbol, orderResponse.Script, orderResponse.OrderID)
				}
				fmt.Println("object ID:", objectId["_id"])
				//Updates Mongo with key executed "false" based on objectId
				db.UpdateMongoAsExecuted(client, objectId, ltp.Ltp, false)
			}
			
			// Buy increase the quantity of the stocks which are performing
			if percentChange > stoploss && data.Executed {
				//Get Balance in the account
				account, err := A.GetRMS()
				if err != nil {
					fmt.Println(err)
				}
				availableFunds, err := strconv.ParseFloat(account.AvailableCash, 64)
				fmt.Println("Available funds are:", availableFunds)

				// Check balance and place order
				if availableFunds >= ltp.Ltp {
					stk := SmartApi.OrderParams{
						Variety: data.Variety,
						TradingSymbol: pos.Tradingsymbol,
						SymbolToken: pos.SymbolToken,
						TransactionType: "BUY",
						Exchange: data.Exchange,
						OrderType: data.OrderType,
						ProductType: data.ProductType,
						Duration: data.Duration,
						Price: ltp.Ltp,
						SquareOff: data.SquareOff,
						StopLoss: data.StopLoss,
						Quantity: "1",
						Executed: true,
					}
					order, err := A.PlaceOrder(stk)
					if err != nil {
						fmt.Println("failed to place repeat order", err)
					}
					fmt.Println("Placed repeat orderer with Order ID and Script :- ", order)
					db.UpdateMongoAsExecuted(client, objectId, ltp.Ltp, true)
				}
			}
			
			session.UserSessionTokens, err = A.RenewAccessToken(session.RefreshToken)
			if err != nil {
				fmt.Println("failed to refresh token:", err)
			}
		}
		loopvar = len(positions)	
	}
}