package httputils

import (
	"bytes"
	"fmt"
	"net/http"
)

// 封装post指令
func Post(url string, requestBody []byte) (http.Response, error) {
	// jsonStr, err := json.Marshal(requestBody)
	// if err != nil {
	// 	fmt.Println("marshal request body failed")
	// 	return
	// }
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("post request failed")
		return http.Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("send post request failed")
		return http.Response{}, err
	}
	defer resp.Body.Close()
	return *resp, nil
}

// 封装get指令
func Get(url string) (http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("get request failed")
		return http.Response{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("send get request failed")
		return http.Response{}, err
	}
	defer resp.Body.Close()
	return *resp, nil
}
