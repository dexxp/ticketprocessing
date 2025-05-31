package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	baseURL = "http://localhost:8080"
)

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type CreateRentalRequest struct {
	EquipmentID uint      `json:"equipment_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Comment     string    `json:"comment"`
}

func (c *Client) Register(name, email, password string) error {
	req := RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}

	resp, err := c.sendRequest("POST", "/register", req, false)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Registration successful!")
	return nil
}

func (c *Client) Login(email, password string) error {
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.sendRequest("POST", "/login", req, false)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.token = authResp.Token
	fmt.Println("Login successful! Token received.")
	return nil
}

func (c *Client) GetMe() error {
	resp, err := c.sendRequest("GET", "/me", nil, true)
	if err != nil {
		return fmt.Errorf("get me failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get me failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Current user info: %+v\n", result)
	return nil
}

func (c *Client) CreateRentalRequest(equipmentID uint, startDate, endDate time.Time, comment string) error {
	req := CreateRentalRequest{
		EquipmentID: equipmentID,
		StartDate:   startDate,
		EndDate:     endDate,
		Comment:     comment,
	}

	resp, err := c.sendRequest("POST", "/rental_request", req, true)
	if err != nil {
		return fmt.Errorf("create rental request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create rental request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Rental request created: %+v\n", result)
	return nil
}

func (c *Client) GetRequestStatus(requestID uint) error {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/rental_request/%d/status", requestID), nil, true)
	if err != nil {
		return fmt.Errorf("get request status failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get request status failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Request status: %+v\n", result)
	return nil
}

func (c *Client) GetRequestStatusAt(requestID uint, datetime time.Time) error {
	url := fmt.Sprintf("/rental_request/%d/status_at?datetime=%s", requestID, datetime.Format(time.RFC3339))
	resp, err := c.sendRequest("GET", url, nil, true)
	if err != nil {
		return fmt.Errorf("get request status at failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get request status at failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Request status at %s: %+v\n", datetime.Format(time.RFC3339), result)
	return nil
}

func (c *Client) sendRequest(method, path string, body interface{}, auth bool) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if auth {
		if c.token == "" {
			return nil, fmt.Errorf("no authentication token available")
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

func main() {
	client := NewClient()

	// Example usage
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  register <name> <email> <password>")
		fmt.Println("  login <email> <password>")
		fmt.Println("  me")
		fmt.Println("  create-request <equipment_id> <start_date> <end_date> <comment>")
		fmt.Println("  get-status <request_id>")
		fmt.Println("  get-status-at <request_id> <datetime>")
		return
	}

	command := os.Args[1]
	var err error

	switch command {
	case "register":
		if len(os.Args) != 5 {
			fmt.Println("Usage: client register <name> <email> <password>")
			return
		}
		err = client.Register(os.Args[2], os.Args[3], os.Args[4])

	case "login":
		if len(os.Args) != 4 {
			fmt.Println("Usage: client login <email> <password>")
			return
		}
		err = client.Login(os.Args[2], os.Args[3])

	case "me":
		err = client.GetMe()

	case "create-request":
		if len(os.Args) != 6 {
			fmt.Println("Usage: client create-request <equipment_id> <start_date> <end_date> <comment>")
			return
		}
		equipmentID := uint(0)
		fmt.Sscanf(os.Args[2], "%d", &equipmentID)
		startDate, _ := time.Parse(time.RFC3339, os.Args[3])
		endDate, _ := time.Parse(time.RFC3339, os.Args[4])
		err = client.CreateRentalRequest(equipmentID, startDate, endDate, os.Args[5])

	case "get-status":
		if len(os.Args) != 3 {
			fmt.Println("Usage: client get-status <request_id>")
			return
		}
		requestID := uint(0)
		fmt.Sscanf(os.Args[2], "%d", &requestID)
		err = client.GetRequestStatus(requestID)

	case "get-status-at":
		if len(os.Args) != 4 {
			fmt.Println("Usage: client get-status-at <request_id> <datetime>")
			return
		}
		requestID := uint(0)
		fmt.Sscanf(os.Args[2], "%d", &requestID)
		datetime, _ := time.Parse(time.RFC3339, os.Args[3])
		err = client.GetRequestStatusAt(requestID, datetime)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		return
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
