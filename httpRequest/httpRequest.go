package httprequest

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	SmartApi "github.com/angel-one/smartapigo"
)

type clientParams struct {
	ClientCode  string `json:"client"`
	Password  string `json:"password"`
	APIKey  string `json:"api_key"`
	TOTPKEY string `json:"totp"`
}

func HttpRequest( url string, method string , payload *strings.Reader, auth clientParams, session SmartApi.UserSession) ([]byte) {
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