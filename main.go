package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {

	url := "https://api.openai.com/v1/chat/completions"
	method := "POST"

	payload := strings.NewReader(`{
    "model": "gpt-3.5-turbo",
    "messages": [
        {
            "role": "user",
            "content": "What is OpenAI?"
        }
    ]
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer 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")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
