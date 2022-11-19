package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"path/filepath"

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
		var parsed NeocitiesErrResp
		var serverMsg string
		err := json.Unmarshal(body, &parsed)
		if err == nil {
			serverMsg = parsed.Message
		} else {
			serverMsg = string(body)
		}
		fmt.Printf("Neocities error %d:\n%s\n", resp.StatusCode, serverMsg)
		return fmt.Errorf(
			"Credentials check failed: [%d] %s", resp.StatusCode, serverMsg,
		)

	}
}

type NeocitiesErrResp struct {
	Message string `json:"message"`
}

func PublishNeocities(src fs.FS, nc *models.Neocities) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := fs.WalkDir(src, ".", func(path string, d fs.DirEntry, e error) error {
		if d.IsDir() {
			return nil
		}

		srcFile, err := src.Open(path)
		if err != nil {
			return fmt.Errorf("open src file: %w", err)
		}
		defer srcFile.Close()

		part, err := writer.CreateFormFile(path, filepath.Base(path))
		if err != nil {
			return fmt.Errorf("create multipart part: %w", err)
		}

		_, err = io.Copy(part, srcFile)
		if err != nil {
			return fmt.Errorf("cp to multipart body: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", NeocitiesBase+"/upload", body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(nc.Username, nc.Password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("calling neocities upload api: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading neocities response: %v", err)
	}

	if resp.StatusCode != 200 {
		var parsed NeocitiesErrResp
		var serverMsg string
		err := json.Unmarshal(respBody, &parsed)
		if err == nil {
			serverMsg = parsed.Message
		} else {
			serverMsg = string(respBody)
		}
		return fmt.Errorf(
			"upload api call error %d:\n%s\n", resp.StatusCode, serverMsg,
		)
	}

	// TODO: prune old files on neocities.

	return nil
}
