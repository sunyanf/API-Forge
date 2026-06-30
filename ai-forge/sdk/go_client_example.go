package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer YOUR_JWT_HERE")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
	}
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}
	fmt.Printf("User profile: %v\n", data)
}
