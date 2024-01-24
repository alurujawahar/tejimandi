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

func calPercentageChange(currentPrice float64, avgTradePrice string) float64 {
	atp, _ := strconv.ParseFloat(avgTradePrice, 64)
	percentageChangeFromAtp := ((currentPrice - atp)/atp)*100
	return percentageChangeFromAtp
}

func calNewATP(ltp float64, presentQuantity string, presentATP string) string {
	quantity, _ := strconv.ParseFloat(presentQuantity, 64)
	ATP, _ := strconv.ParseFloat(presentATP, 64)
	// return (ltp + quantity*ATP)/(quantity + 1)
	s := fmt.Sprintf("%.2f", (ltp + quantity*ATP)/(quantity + 1))
	return s
}

func MonitorOrders(A *SmartApi.Client, auth h.ClientParams, session SmartApi.UserSession, client *mongo.Client) {
	loopvar := 1
	for loopvar != 0 {
		positions, err := A.GetPositions()
		if err != nil {
			fmt.Println("Error getting your Positions", err)
			os.Exit(1)
		}
		fmt.Println("length of Positions", len(positions))
		for _, pos := range positions {
			fmt.Println("##############################################################################\n")
			if err != nil {
				fmt.Println("err:", err)
			}
			if pos.ProductType == "INTRADAY" && pos.NetQty != "0" {
				fmt.Printf("Symbol: %s : ATP Value: %s, Net Price: %s\n", pos.Tradingsymbol, pos.AverageNetPrice, pos.NetPrice)
			}

			// ltpPercentageChange := getValueChange(pos.SymbolToken, pos.Tradingsymbol, auth, session)
			ltpparams := SmartApi.LTPParams{
				Exchange: pos.Exchange,
				SymbolToken: pos.SymbolToken,
				TradingSymbol: pos.Tradingsymbol,
			}
			ltp, err := A.GetLTP(ltpparams)
			if err != nil {
				fmt.Println("unable to get Ltp for:", pos.Tradingsymbol, ltp)
			}
			fmt.Println("LTP:", ltp.Ltp)	

			percentChange := calPercentageChange(ltp.Ltp, pos.AverageNetPrice)
			fmt.Printf("percentage change of %s is %v \n:", pos.Tradingsymbol, percentChange)

			data, objectId := db.QueryMongo(client, pos.Tradingsymbol)

			// if !(data.Executed == false && pos.NetQty == "0") {
			// 	fmt.Errorf("There is a mismatch of the Quantity with posistion and Data", pos.Tradingsymbol)
			// 	continue
			// }
			//Sell stocks if they are less than stoploss
			if percentChange < stoploss && data.Executed {
				exitParams := SmartApi.OrderParams{
					Exchange: pos.Exchange,
					Variety: "NORMAL",
					TradingSymbol: pos.Tradingsymbol,
					SymbolToken: pos.SymbolToken,
					OrderType: "MARKET",
					ProductType: "INTRADAY",
					Duration: "DAY",
					SquareOff: "0",
					StopLoss: "0",
					Quantity: pos.NetQty,
					TransactionType: "SELL",
					Executed: false,
				}
				if true {
					orderResponse, err := A.PlaceOrder(exitParams)
					if err != nil {
						fmt.Println("Failed to exit position", err)
					}
					fmt.Println("Successfully exited trading Symbol", pos.Tradingsymbol, orderResponse.Script, orderResponse.OrderID)
				
					fmt.Printf("object ID of Symbol %s is %s:", pos.Tradingsymbol, objectId["_id"])
					//Updates Mongo with key executed "false" based on objectId
					db.UpdateMongoAsExecuted(client, objectId, ltp.Ltp, false, "0")
				}
			}
			//Calculate New ATP based on present LTP
			newATP := calNewATP(ltp.Ltp, pos.NetQty, pos.AverageNetPrice)
			percentChangeWithNewATP := calPercentageChange(ltp.Ltp, newATP)
			fmt.Println("Percentage Change with new ATP is: ", percentChangeWithNewATP)

			// Buy increase the quantity of the stocks which are performing
			if (percentChangeWithNewATP > percentChange && percentChangeWithNewATP > 0   && percentChange > stoploss && data.Executed)  {
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
						OrderType: "MARKET",
						ProductType: data.ProductType,
						Duration: data.Duration,
						SquareOff: data.SquareOff,
						StopLoss: data.StopLoss,
						Quantity: "1",
						Executed: true,
					}
					if true {
						order, err := A.PlaceOrder(stk)
						if err != nil {
							fmt.Println("failed to place repeat order", err)
						}
						fmt.Println("Placed repeat orderer with Order ID and Script :- ", order)
						OldQuantity, err := strconv.ParseInt(pos.NetQty, 10, 64)
						if err != nil {
							fmt.Println("Error:", err)
						}
						newQuantity := fmt.Sprint(OldQuantity + 1)
						db.UpdateMongoAsExecuted(client, objectId, ltp.Ltp, true, newQuantity)
					}
				}
			}
			
			session.UserSessionTokens, err = A.RenewAccessToken(session.RefreshToken)
			if err != nil {
				fmt.Println("failed to refresh token:", err)
			}
			fmt.Println("##############################################################################\n")
		}
		loopvar = len(positions)	
	}
}