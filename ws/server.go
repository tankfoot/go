// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"
	"html/template"
	"log"
    "fmt"
	"net/http"
    "encoding/json"
    "os/exec"
    "errors"
    "io/ioutil"
    "bytes"

	"github.com/gorilla/websocket"
    sj "github.com/bitly/go-simplejson"
)

//Incoming Json struct
type Data struct {
    Query string
}

type Message struct {
    Header [6]float64
    Data   Data
}

//Dialogflow Query struct
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

//var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func DetectIntentText(projectID, sessionID, text, languageCode string) (string, error) {
    if projectID == "" || sessionID == "" {
        return "", errors.New(fmt.Sprintf("Received empty project (%s) or session (%s)", projectID, sessionID))
    }
    basePath := "https://dialogflow.googleapis.com/v2/"
    sessionPath := fmt.Sprintf("projects/%s/agent/sessions/%s", projectID, sessionID)

    client := &http.Client{}
    var jsonData QueryInput
    jsonData.QueryInput = TextInput{Text{Text: text, LanguageCode: languageCode}}
    jsonValue, _ := json.Marshal(jsonData)
    detectIntentUrl := basePath + sessionPath + ":detectIntent"
    r, err := http.NewRequest("POST", detectIntentUrl, bytes.NewBuffer(jsonValue))

    token, _ := GetGcloudToken()
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
        return r, nil
    }

    return "", nil
}

func GetGcloudToken() (string, error) {
    cmd := exec.Command("gcloud", 
        "auth",
        "application-default",
        "print-access-token",)

    out, err := cmd.Output()
    if err != nil {
        log.Fatal(err)
        return "", err
    }

    token := string(out)[:len(string(out))-1] // line ending subtract
    return token, nil
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

        var m Message
        err1 := json.Unmarshal(message, &m)
        if err1 != nil {
            log.Fatalln("error:", err1)
        }
        rr, _ := DetectIntentText("chipotle-aeeb4", "123", m.Data.Query, "en")
		err = c.WriteMessage(mt, []byte(rr))

		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo/")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	//log.Fatal(http.ListenAndServe(*addr, nil))
    log.Fatal(http.ListenAndServe(":8080", nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.innerHTML = message;
        output.appendChild(d);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))
