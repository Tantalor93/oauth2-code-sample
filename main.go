package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Token struct {
	AccessToken string `json:"access_token"`
}

func main() {

	setUpLogging()

	clientId, clientSecret := funcName()

	router := mux.NewRouter()
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "https://accounts.google.com/o/oauth2/v2/auth?"+
			"scope=https://www.googleapis.com/auth/drive.metadata.readonly&access_type=offline&"+
			"include_granted_scopes=true&state=state_parameter_passthrough_value&redirect_uri=http://localhost:8081/callback&"+
			"response_type=code&client_id="+*clientId, 301)
	})

	router.HandleFunc("/callback", func(writer http.ResponseWriter, request *http.Request) {
		get := request.URL.Query().Get("code")
		log.WithField("code", get).Info("code received")

		data := url.Values{}
		data.Set("client_id", *clientId)
		data.Add("client_secret", *clientSecret)
		data.Add("redirect_uri", "http://localhost:8081/callback")
		data.Add("grant_type", "authorization_code")
		data.Add("code", get)

		r, _ := http.NewRequest(
			"POST",
			"https://www.googleapis.com/oauth2/v4/token",
			strings.NewReader(data.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		defer request.Body.Close()
		resp, err := client.Do(r)
		if err != nil {
			log.Error(err)
			return
		}
		var token Token
		if resp.StatusCode == 200 {
			json.NewDecoder(resp.Body).Decode(&token)
			defer resp.Body.Close()
			log.WithField("access_token", token.AccessToken).Info("access_token received")

		} else {
			log.WithField("status", resp.StatusCode).Error("unexpected status")
		}
	})

	http.ListenAndServe(":8081", router)

}

func funcName() (*string, *string) {
	clientId := flag.String("client_id", "", "OAUTH2 client_id")
	clientSecret := flag.String("client_secret", "", "OAUTH2 client_secret")

	flag.Parse()

	if *clientId == "" {
		panic("Please provided --client_id flag")
	}

	if *clientSecret == "" {
		panic("Please provided --client_secret flag")
	}

	return clientId, clientSecret
}

func setUpLogging() {
	formatter := &log.JSONFormatter{}
	log.SetFormatter(formatter)
}

