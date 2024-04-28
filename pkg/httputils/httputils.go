package httputils

import (
	"bytes"
	"fmt"
	"net/http"
)

// 封装post指令
func Post(url string, requestBody []byte) {
	// jsonStr, err := json.Marshal(requestBody)
	// if err != nil {
	// 	fmt.Println("marshal request body failed")
	// 	return
	// }
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("post request failed")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("send post request failed")
		return
	}
	defer resp.Body.Close()
}

// 封装get指令
func Get(url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("get request failed")
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("send get request failed")
		return
	}
	defer resp.Body.Close()
}
