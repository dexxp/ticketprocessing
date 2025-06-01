package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	baseURL = "http://localhost:8080"
)

type Session struct {
	Token      string
	UserInfo   map[string]interface{}
	IsLoggedIn bool
}

type Client struct {
	httpClient *http.Client
	session    *Session
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		session: &Session{
			UserInfo:   make(map[string]interface{}),
			IsLoggedIn: false,
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

type Equipment struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	AvailableQuantity int    `json:"available_quantity"`
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

	c.session.Token = authResp.Token
	c.session.IsLoggedIn = true

	// Get user info after login
	if err := c.GetMe(); err != nil {
		return fmt.Errorf("failed to get user info after login: %w", err)
	}

	fmt.Println("Login successful! You are now logged in.")
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

	if err := json.NewDecoder(resp.Body).Decode(&c.session.UserInfo); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

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

func (c *Client) CreateEquipment(name string, quantity int) error {
	req := Equipment{
		Name:              name,
		AvailableQuantity: quantity,
	}

	resp, err := c.sendRequest("POST", "/api/equipment", req, true)
	if err != nil {
		return fmt.Errorf("create equipment failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create equipment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var equipment Equipment
	if err := json.NewDecoder(resp.Body).Decode(&equipment); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Equipment created: %+v\n", equipment)
	return nil
}

func (c *Client) GetEquipment(id uint) error {
	resp, err := c.sendRequest("GET", fmt.Sprintf("/api/equipment/%d", id), nil, true)
	if err != nil {
		return fmt.Errorf("get equipment failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get equipment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var equipment Equipment
	if err := json.NewDecoder(resp.Body).Decode(&equipment); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Equipment: %+v\n", equipment)
	return nil
}

func (c *Client) UpdateEquipment(id uint, name string, quantity int) error {
	req := Equipment{
		ID:                id,
		Name:              name,
		AvailableQuantity: quantity,
	}

	resp, err := c.sendRequest("PUT", fmt.Sprintf("/api/equipment/%d", id), req, true)
	if err != nil {
		return fmt.Errorf("update equipment failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update equipment failed with status %d: %s", resp.StatusCode, string(body))
	}

	var equipment Equipment
	if err := json.NewDecoder(resp.Body).Decode(&equipment); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Equipment updated: %+v\n", equipment)
	return nil
}

func (c *Client) DeleteEquipment(id uint) error {
	resp, err := c.sendRequest("DELETE", fmt.Sprintf("/api/equipment/%d", id), nil, true)
	if err != nil {
		return fmt.Errorf("delete equipment failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete equipment failed with status %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Equipment %d deleted successfully\n", id)
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
		if !c.session.IsLoggedIn || c.session.Token == "" {
			return nil, fmt.Errorf("no authentication token available")
		}
		req.Header.Set("Authorization", "Bearer "+c.session.Token)
	}

	return c.httpClient.Do(req)
}

func printWelcome() {
	fmt.Println("\n=== Welcome to the Rental Equipment Client ===")
	fmt.Println("Type 'help' to see available commands")
	fmt.Println("Type 'exit' to quit the application")
	fmt.Println("===========================================\n")
}

func printHelp(isLoggedIn bool) {
	fmt.Println("\nAvailable commands:")
	if !isLoggedIn {
		fmt.Println("  register <name> <email> <password> - Register a new account")
		fmt.Println("  login <email> <password>          - Login to your account")
	} else {
		fmt.Println("  me                                - Show your profile information")
		fmt.Println("\nEquipment Management:")
		fmt.Println("  create-equipment <name> <quantity> - Create new equipment")
		fmt.Println("  get-equipment <id>                - Get equipment details")
		fmt.Println("  update-equipment <id> <name> <quantity> - Update equipment")
		fmt.Println("  delete-equipment <id>             - Delete equipment")
		fmt.Println("\nRental Requests:")
		fmt.Println("  create-request <equipment_id> <start_date> <end_date> <comment> - Create a rental request")
		fmt.Println("  get-status <request_id>           - Get status of a rental request")
		fmt.Println("  get-status-at <request_id> <datetime> - Get status of a rental request at specific time")
		fmt.Println("  logout                            - Logout from your account")
	}
	fmt.Println("  help                               - Show this help message")
	fmt.Println("  exit                               - Exit the application")
	fmt.Println()
}

func (c *Client) Logout() {
	c.session.Token = ""
	c.session.IsLoggedIn = false
	c.session.UserInfo = make(map[string]interface{})
	fmt.Println("Successfully logged out!")
}

func parseCommandLine(input string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		switch input[i] {
		case '"':
			if i > 0 && input[i-1] == '\\' {
				// Handle escaped quotes
				current.WriteByte(input[i])
			} else {
				inQuotes = !inQuotes
			}
		case ' ':
			if inQuotes {
				current.WriteByte(input[i])
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(input[i])
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

func main() {
	client := NewClient()
	scanner := bufio.NewScanner(os.Stdin)

	printWelcome()
	printHelp(client.session.IsLoggedIn)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if len(input) == 0 {
			continue
		}

		args := parseCommandLine(input)
		if len(args) == 0 {
			continue
		}

		command := args[0]
		var err error

		switch command {
		case "help":
			printHelp(client.session.IsLoggedIn)

		case "exit":
			fmt.Println("Goodbye!")
			return

		case "register":
			if len(args) != 4 {
				fmt.Println("Usage: register <name> <email> <password>")
				continue
			}
			err = client.Register(args[1], args[2], args[3])
			if err == nil {
				fmt.Println("Registration successful! Please login to continue.")
			}

		case "login":
			if len(args) != 3 {
				fmt.Println("Usage: login <email> <password>")
				continue
			}
			err = client.Login(args[1], args[2])
			if err == nil {
				printHelp(true)
			}

		case "logout":
			if !client.session.IsLoggedIn {
				fmt.Println("You are not logged in!")
				continue
			}
			client.Logout()
			printHelp(false)

		case "me":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			err = client.GetMe()
			if err == nil {
				fmt.Printf("Your profile: %+v\n", client.session.UserInfo)
			}

		case "create-equipment":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			if len(args) != 3 {
				fmt.Println("Usage: create-equipment <name> <quantity>")
				continue
			}
			quantity := 0
			fmt.Sscanf(args[2], "%d", &quantity)
			err = client.CreateEquipment(args[1], quantity)

		case "get-equipment":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			if len(args) != 2 {
				fmt.Println("Usage: get-equipment <id>")
				continue
			}
			id := uint(0)
			fmt.Sscanf(args[1], "%d", &id)
			err = client.GetEquipment(id)

		case "update-equipment":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			if len(args) != 4 {
				fmt.Println("Usage: update-equipment <id> <name> <quantity>")
				continue
			}
			id := uint(0)
			quantity := 0
			fmt.Sscanf(args[1], "%d", &id)
			fmt.Sscanf(args[3], "%d", &quantity)
			err = client.UpdateEquipment(id, args[2], quantity)

		case "delete-equipment":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			if len(args) != 2 {
				fmt.Println("Usage: delete-equipment <id>")
				continue
			}
			id := uint(0)
			fmt.Sscanf(args[1], "%d", &id)
			err = client.DeleteEquipment(id)

		case "create-request":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			if len(args) != 5 {
				fmt.Println("Usage: create-request <equipment_id> <start_date> <end_date> <comment>")
				fmt.Println("Example: create-request 1 \"2024-03-20T10:00:00Z\" \"2024-03-25T18:00:00Z\" \"Need for weekend\"")
				continue
			}
			equipmentID := uint(0)
			fmt.Sscanf(args[1], "%d", &equipmentID)
			startDate, err := time.Parse(time.RFC3339, args[2])
			if err != nil {
				fmt.Println("Invalid start date format. Use RFC3339 format (e.g., 2024-03-20T10:00:00Z)")
				continue
			}
			endDate, err := time.Parse(time.RFC3339, args[3])
			if err != nil {
				fmt.Println("Invalid end date format. Use RFC3339 format (e.g., 2024-03-20T10:00:00Z)")
				continue
			}
			err = client.CreateRentalRequest(equipmentID, startDate, endDate, args[4])

		case "get-status":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			if len(args) != 2 {
				fmt.Println("Usage: get-status <request_id>")
				continue
			}
			requestID := uint(0)
			fmt.Sscanf(args[1], "%d", &requestID)
			err = client.GetRequestStatus(requestID)

		case "get-status-at":
			if !client.session.IsLoggedIn {
				fmt.Println("Please login first!")
				continue
			}
			if len(args) != 3 {
				fmt.Println("Usage: get-status-at <request_id> <datetime>")
				continue
			}
			requestID := uint(0)
			fmt.Sscanf(args[1], "%d", &requestID)
			datetime, err := time.Parse(time.RFC3339, args[2])
			if err != nil {
				fmt.Println("Invalid datetime format. Use RFC3339 format (e.g., 2024-03-20T15:04:05Z)")
				continue
			}
			err = client.GetRequestStatusAt(requestID, datetime)

		default:
			fmt.Printf("Unknown command: %s\nType 'help' to see available commands\n", command)
			continue
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}
