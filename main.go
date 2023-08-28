package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8025/healthz", "URL liveness probe")
	urlNotif := flag.String("url_notif", "http://localhost:8025/api/v1/notif/telegram", "URL Notif Telegram")
	micro := flag.String("micro", "micro-mms-qris", "Name Micro")
	chatID := flag.Int64("chat_id", 0, "Chat ID Telegram")
	flag.Parse()
	ctx := context.Background()
	// Melakukan HTTP GET request
	_, _, err := client_Req(ctx, []headers{}, *url, "GET", []byte{}, 100)
	if err != nil {
		errTeleg := sendTelegramNotifError(ctx, *urlNotif, *chatID, fmt.Sprintf("Halo harap check monitoring queue anda karena terjadi kesalahan di micro :  <b>%s</b> , dengan message : <b>%s</b> , dan data sebagai berikut :\n%v\n", *micro, "Liveness probe gagal", err.Error()))
		fmt.Println(errTeleg)
		fmt.Printf("Liveness probe gagal: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Liveness probe sukses: %s\n", *url)
}

type headers struct {
	Key   string
	Value string
}

func sendTelegramNotifError(ctx context.Context, url string, chat_id int64, msg string) error {
	headers := []headers{
		{
			Key:   "Content-Type",
			Value: "Application/json",
		},
	}
	var chatID int64
	if chat_id == 0 {
		chatID = -932412566
	}

	requestData := map[string]interface{}{
		"chat_id": chatID,
		"message": msg,
	}

	bodyBytes, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	_, _, err = client_Req(ctx, headers, url, "POST", bodyBytes, 60)

	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return nil
}

func client_Req(ctx context.Context, h []headers, uri string, m string, d []byte, timeout int64) ([]byte, http.Header, error) {

	client := &http.Client{Timeout: time.Second * time.Duration(timeout)}
	req, err := http.NewRequest(m, uri, bytes.NewReader(d))
	if err != nil {
		return nil, nil, err
	}

	if h == nil {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/json")
		//header tambahan
		for i := range h {
			req.Header[h[i].Key] = []string{h[i].Value}
		}
	}

	// reqHeader, _ := json.Marshal(req.Header)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(err)
	if err != nil {
		return nil, nil, err
	}

	//check response status
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= http.StatusInternalServerError {
			return nil, nil, errors.New(fmt.Sprintf("%d error: %s", resp.StatusCode, resp.Status))
		}
		return nil, nil, fmt.Errorf("%d error: %s", resp.StatusCode, resp.Status)
	}

	return body, resp.Header, nil
}

