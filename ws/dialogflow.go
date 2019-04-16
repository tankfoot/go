package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os/exec"
    "log"

    sj "github.com/bitly/go-simplejson"
)

type Text struct {
		Text string `json:"text"`
		LanguageCode string `json:"languageCode"`
}

type TextInput struct {
		TextInput Text `json:"text"`
}

type QueryInput struct {
	QueryInput TextInput `json:"queryInput"`
}

func main() {
    cmd := exec.Command("gcloud", 
		"auth",
		"application-default",
		"print-access-token",)

	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

    apiUrl := "https://dialogflow.googleapis.com/v2/projects/container-a3c3c/agent/sessions/123:detectIntent"

    client := &http.Client{}
    var jsonData QueryInput
    jsonData.QueryInput = TextInput{Text{Text: "Hello", LanguageCode: "en"}}
    jsonValue, _ := json.Marshal(jsonData)
    fmt.Println(string(jsonValue))
    r, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonValue)) // URL-encoded payload
 
    token := string(out)[:len(string(out))-1] // line ending subtract
    var bearer = "Bearer " + token
    r.Header.Add("Authorization", bearer)
    r.Header.Add("Content-Type", "application/json; charset=utf-8")

    resp, err := client.Do(r)

    if err != nil {
        fmt.Printf("The HTTP request failed with error %s\n", err)
    } else {
        data, _ := ioutil.ReadAll(resp.Body)
        js, _ := sj.NewJson(data)
        r := js.Get("queryResult").Get("fulfillmentText").MustString()
        fmt.Printf("%s\n", r)
    }

}