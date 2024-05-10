package httputils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func ParseResponse(response *http.Response) []byte {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	return body
}

// 封装post指令
func Post(url string, requestBody []byte) []byte {
	// jsonStr, err := json.Marshal(requestBody)
	// if err != nil {
	// 	fmt.Println("marshal request body failed")
	// 	return
	// }
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("post request failed")
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("send post request failed")
		return nil
	}
	defer resp.Body.Close()
	return ParseResponse(resp)
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
