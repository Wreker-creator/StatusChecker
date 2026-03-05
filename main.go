package main

import (
	"context"
	"fmt"
	"net/http"
	requesthandler "statusChecker/requestHandler"
	"time"
)

func main() {

	client := &http.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results := requesthandler.CheckURLs(ctx, []string{"https://google.com"}, client)

	for url, result := range results {
		fmt.Println("URL:", url)
		fmt.Println("Status:", result.StatusCode)
		fmt.Println("Error:", result.Err)
	}

}
