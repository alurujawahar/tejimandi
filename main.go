// token := tokenLookUp("HDFCBANK", instrument_list, "NSE" )
// fmt.Println(token)

// symbol := symbolLookUp(token, instrument_list, "NSE")
// fmt.Println(symbol)

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	SmartApi "github.com/angel-one/smartapigo"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
const stoploss = -0.2

const mongoUrl = "mongodb://localhost:27017"
 
type clientParams struct {
	ClientCode  string `json:"client"`
	Password  string `json:"password"`
	APIKey  string `json:"api_key"`
	TOTPKEY string `json:"totp"`
}

type Instrument struct {
	Token string `json:"token"`
	Symbol string `json:"symbol"`
	Name string `json:"name"`
	Expiry string `json:"expiry"`
	Strike string `json:"strike"`
	Lotsize string `json:"lotsize"`
	Instrumenttype string `json:"instrumenttype"`
	Exch_seg string `json:"exch_seg"`
	Tick_size string `json:"tick_size"`
}

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

func httpRequest( url string, method string , payload *strings.Reader,  auth clientParams, session SmartApi.UserSession) ([]byte) {
	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
	fmt.Println(err)
	os.Exit(1)
	}
	req.Header.Add("X-PrivateKey", auth.APIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-SourceID", "WEB")
	req.Header.Add("X-ClientLocalIP", "CLIENT_LOCAL_IP")
	req.Header.Add("X-ClientPublicIP", "CLIENT_PUBLIC_IP")
	req.Header.Add("X-MACAddress", "MAC_ADDRESS")
	req.Header.Add("X-UserType", "USER")
	req.Header.Add("Authorization", "Bearer "+session.AccessToken)
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

	return body 
}

func tokenLookUp(ticker string , instrument_list []Instrument, exchange string)  string  {
	var foundToken string
	for _, inst := range instrument_list {
		if inst.Symbol == ticker && inst.Exch_seg == exchange && strings.Split(inst.Symbol, "-")[1] == "EQ"{
			foundToken = inst.Token
		}
	}	
	return foundToken
}

// func symbolLookUp(token string, instrument_list []Instrument, exchange string)  Instrument  {
// 	var foundSymbol Instrument
// 	for _, inst := range instrument_list {
// 		if inst.Token == token && inst.Exch_seg == exchange && strings.Split(inst.Symbol, "-")[1] == "EQ"{
// 			foundSymbol = inst
// 		}
// 	}	
// 	return foundSymbol
// }

func orderBook(A *SmartApi.Client, auth clientParams, session SmartApi.UserSession) {
	url := "https://apiconnect.angelbroking.com/rest/secure/angelbroking/order/v1/getTradeBook"
	method := "GET"
	var payload *strings.Reader
	body := httpRequest(url, method, payload, auth, session)
	fmt.Println("Orders: ", string(body))
}

func getInstrumentList() ([]Instrument) {
	const instrument_url = "https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json"
	var instrument_list []Instrument
	response, err := http.Get(instrument_url)
	if err != nil {
		fmt.Println("Error opening instrument list url %v", err)
	}
	instrument_byte, _ := io.ReadAll(response.Body)

	if json.Unmarshal(instrument_byte, &instrument_list) != nil {
		fmt.Println("Unable to unMarshal response %v", err)
	}
	token := tokenLookUp("YESBANK", instrument_list, "NSE" )
	fmt.Println(token)
	return instrument_list
}

func getValueChange(token string, symbol string, auth clientParams, session SmartApi.UserSession) float64 {
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
	body := httpRequest(url, method, payload, auth, session)

	json.Unmarshal(body, &positionData)
	// symbol := symbolLookUp(token, instrument_list, "NSE")
		for _, f := range positionData.Data.Fetched {
			fmt.Printf("Pecentage change of token %s is %.2f\n", symbol , f.PercentChange)
			percentageChange = f.PercentChange
		}
	
	return percentageChange
	
}

func queryMongo(client *mongo.Client, tradingSymbol string) (SmartApi.OrderParams, bson.M){
	// Access a MongoDB collection
	collection := client.Database("stocks").Collection("list")

	// Define a filter for the query
	filter := bson.D{{Key: "tradingsymbol", Value: bson.D{{Key: "$eq", Value: tradingSymbol}}}}
	options := options.FindOne().SetProjection(bson.D{{Key: "_id", Value: 1}})

	// Define options for the query (e.g., sorting)
	// options := options.FindOne().SetSort(bson.D{{Key: "age", Value: 1}})
	// Find a single document in the collection based on the filter and options
	var objectId bson.M
	var result SmartApi.OrderParams
	err := collection.FindOne(context.Background(), filter, options).Decode(&objectId)
	if err == mongo.ErrNoDocuments {
		fmt.Println("No matching document found.")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Query executed successfully!")
		fmt.Println(objectId["_id"])
	}

	err = collection.FindOne(context.Background(), filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		fmt.Println("No matching document found.")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Query executed successfully!")
	}
	return result, objectId
}

func updateMongo(client *mongo.Client, _id bson.M) {
	collection := client.Database("stocks").Collection("list")

    // Define the filter based on the document's _id
    filter := bson.D{{Key: "_id", Value: _id["_id"]}} // Replace with the actual _id

    // Define the update to be performed
    update := bson.D{
        {Key: "$set", Value: bson.D{{Key: "executed", Value: false}}}, 
    }

    // Perform the update
    result, err := collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Matched %v document(s) and modified %v document(s)\n", result.MatchedCount, result.ModifiedCount)

}

func monitorOrders(A *SmartApi.Client, auth clientParams, session SmartApi.UserSession, client *mongo.Client) {
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
			percentChange := getValueChange(pos.SymbolToken, pos.Tradingsymbol, auth, session)
			ltpparams.Exchange = pos.Exchange
			ltpparams.SymbolToken = pos.SymbolToken
			ltpparams.TradingSymbol = pos.Tradingsymbol
			ltp, err := A.GetLTP(ltpparams)
			if err != nil {
				fmt.Println("unable to get Ltp for:", pos.Tradingsymbol, ltp)
			}
			data, objectId := queryMongo(client, pos.Tradingsymbol)
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
				if true {
					orderResponse, err := A.PlaceOrder(exitParams)
					if err != nil {
						fmt.Println("Failed to exit position", err)
					}
					fmt.Println("Successfully exited trading Symbol", pos.Tradingsymbol, orderResponse.Script, orderResponse.OrderID)
				}
				fmt.Println("object ID:", objectId["_id"])
				updateMongo(client, objectId)
			}
			session.UserSessionTokens, err = A.RenewAccessToken(session.RefreshToken)
			if err != nil {
				fmt.Println("failed to refresh token:", err)
			}
		}
		loopvar = len(positions)	
	}
}

// func updateJson(OrderParams []SmartApi.OrderParams, tradingsymbol string, latestprice float64) ([]SmartApi.OrderParams){

// 	for i := range OrderParams {
// 		if OrderParams[i].TradingSymbol == tradingsymbol {
// 			OrderParams[i].Price = latestprice
// 		}
// 	}

// 	return OrderParams
// }

func placeBulkOrder(A *SmartApi.Client, s string, exchange string, client *mongo.Client)  {
	var OrderParams []SmartApi.OrderParams
	var ltpParams SmartApi.LTPParams
	// instrument_list := getInstrumentList()
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
		// token := tokenLookUp(stk.TradingSymbol, instrument_list, exchange)
		// fmt.Println(stk.TradingSymbol, token)
		// stk.SymbolToken = token
		ltpParams.Exchange = exchange
		ltpParams.SymbolToken = stk.SymbolToken
		ltpParams.TradingSymbol = stk.TradingSymbol
		ltpResp, err := A.GetLTP(ltpParams)
		if err != nil {
			fmt.Println(err)
		}
		stk.Price = ltpResp.Ltp

		if true {
			fmt.Println("Placing Order for Stock: ", stk.TradingSymbol)
			order, err := A.PlaceOrder(stk)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Placed Order ID and Script :- ", order)
			stk.Executed = true
		}

		collection := client.Database("stocks").Collection("list")
		_, err = collection.InsertOne(context.Background(), stk)
		if err != nil {
			log.Fatal(err)
		}
		// updatedJson := updateJson(OrderParams, stk.TradingSymbol, ltpResp.Ltp)

		// updateJSON, err := json.MarshalIndent(updatedJson, "", "   ")
		// if err != nil {
		// 	fmt.Println("Error marshalling JSON:", err)
		// 	return
		// }

		// err = ioutil.WriteFile(s, updateJSON, 0655)
		// if err != nil {
		// 	fmt.Println("Error writing JSON file:", err)
		// 	return
		// }	
	}
}

func authenticate(f string) (*SmartApi.Client, clientParams, SmartApi.UserSession) {
	var params clientParams
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

func connectMongo() *mongo.Client {
	clientOptions := options.Client().ApplyURI(mongoUrl)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return client 

}

func main() {
	stocksFilePath := "/Users/alurujawahar/Desktop/angel/tejimandi/stocks.json"
	filepath := "/Users/alurujawahar/Desktop/angel/tejimandi/keys.json"
	placeorder := true

	client := connectMongo()
	//Get Authenticated
	ABClient, authParams, session := authenticate(filepath)

	//Place Bulk Order
	if placeorder {
		placeBulkOrder(ABClient, stocksFilePath, "NSE", client)
	}

	if true {
		monitorOrders(ABClient, authParams, session, client)
	}

	if false {
		orderBook(ABClient, authParams, session)
	}
	if false {
		instrument_list := getInstrumentList()
		token := tokenLookUp("HDFCBANK", instrument_list, "NSE" )
		fmt.Println(token)
	}
}