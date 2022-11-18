package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.imnhan.com/bloghead/models"
)

const NeocitiesBase = "https://neocities.org/api"

func CheckNeocitiesCreds(nc *models.Neocities) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", NeocitiesBase+"/info", nil)
	if err != nil {
		return fmt.Errorf("check neocities creds: %v", err)
	}
	req.SetBasicAuth(nc.Username, nc.Password)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("check neocities creds: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("check neocities creds: %v", err)
	}

	switch resp.StatusCode {
	case 200:
		return nil
	case 403:
		fmt.Printf("Neocities error %d:\n%s\n", resp.StatusCode, body)
		return errors.New("Neocities rejected the new credentials.")
	default:
		fmt.Printf("Neocities error %d:\n%s\n", resp.StatusCode, body)
		return fmt.Errorf(
			"Neocities credentials check failed: %d:\n%s", resp.StatusCode, body,
		)
	}
}
