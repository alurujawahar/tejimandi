package token

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)


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


func TokenLookUp(ticker string , instrument_list []Instrument, exchange string)  string  {
	var foundToken string
	for _, inst := range instrument_list {
		if inst.Symbol == ticker && inst.Exch_seg == exchange && strings.Split(inst.Symbol, "-")[1] == "EQ"{
			foundToken = inst.Token
		}
	}	
	return foundToken
}


func GetInstrumentList() ([]Instrument) {
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
	token := TokenLookUp("YESBANK", instrument_list, "NSE" )
	fmt.Println(token)
	return instrument_list
}