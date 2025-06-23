package huggingface

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type HuggingFaceClient struct {
	apiKey string
}

func New(apiKey string) *HuggingFaceClient {
	return &HuggingFaceClient{
		apiKey: apiKey,
	}
}

func (hf *HuggingFaceClient) GenerateWithHuggingFace(text string) ([]byte, error) {
	op := "clients.huggingface.generatewithhuggingface"
	url := "https://api-inference.huggingface.co/models/stabilityai/stable-diffusion-xl-base-1.0"
	// url := "https://api-inference.huggingface.co/models/stabilityai/stable-diffusion-3-medium-diffusers"

	prompt := fmt.Sprintf(
		"Создай красивое изображение прогноза погоды. Стиль: плоский дизайн, пастельные тона. %s",
		text)

	req, err := http.NewRequest(
		"POST", url,
		bytes.NewBuffer([]byte(fmt.Sprintf(`{""text":"%s"}`, prompt))),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", hf.apiKey))

	// Do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s API returned %d: %s", op, resp.StatusCode, string(body))
	}

	// Reading bin data
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return imageBytes, nil
}
