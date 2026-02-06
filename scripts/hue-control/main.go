package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the Hue Bridge connection details
type Config struct {
	BridgeIP string
	APIKey   string
}

// Group represents a Hue group (room/zone)
type Group struct {
	Name   string     `json:"name"`
	Type   string     `json:"type"`
	Lights []string   `json:"lights"`
	Action GroupState `json:"action"`
}

// GroupState represents the state of a group
type GroupState struct {
	On  bool `json:"on"`
	Bri int  `json:"bri,omitempty"`
	Hue int  `json:"hue,omitempty"`
	Sat int  `json:"sat,omitempty"`
}

// ColorPreset maps color names to Hue and Saturation values
var ColorPresets = map[string][2]int{
	// Format: "name": {hue (0-65535), saturation (0-254)}
	"red":    {0, 254},
	"orange": {5000, 254},
	"yellow": {10000, 254},
	"green":  {25500, 254},
	"cyan":   {35000, 254},
	"blue":   {46920, 254},
	"purple": {50000, 254},
	"pink":   {56100, 254},
	"warm":   {8000, 200}, // Warm white
	"cool":   {34000, 50}, // Cool white
	"white":  {0, 0},      // Pure white (no color)
}

// Light represents a Hue light
type Light struct {
	Name  string     `json:"name"`
	State LightState `json:"state"`
}

// LightState represents the state of a light
type LightState struct {
	On  bool `json:"on"`
	Bri int  `json:"bri,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "setup":
		runSetup()
	case "list":
		runList()
	case "set":
		runSet()
	case "on":
		runOn()
	case "off":
		runOff()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Hue Control - Philips Hue Light Controller

Usage:
  hue-control <command> [options]

Commands:
  setup       Configure Hue Bridge connection (saves to .env)
  list        List available rooms/groups
  set         Set brightness for lights
  on          Turn all lights on
  off         Turn all lights off
  help        Show this help message

Set Command Options:
  --room <name>        Room name to control (default: "all")
  --brightness <0-100> Brightness percentage (default: 100)
  --hue <0-65535>      Hue value for color (optional)
  --sat <0-254>        Saturation value for color (optional)
  --color <name>       Color preset: red, orange, yellow, green, cyan, blue, purple, pink, warm, cool, white

Configuration:
  Authentication defaults to reading from a .env file or environment variables:
  - HUE_BRIDGE_IP
  - HUE_API_KEY

Examples:
  hue-control setup
  hue-control list
  hue-control set --brightness 50
  hue-control set --room "Living Room" --brightness 75
  hue-control set --color blue
  hue-control set --room "Bedroom" --color warm --brightness 60`)
}

func loadConfig() (*Config, error) {
	// Try loading from .env file, but don't fail if it doesn't exist
	// (we might be using system env vars)
	_ = godotenv.Load()

	bridgeIP := os.Getenv("HUE_BRIDGE_IP")
	apiKey := os.Getenv("HUE_API_KEY")

	// Fallback to legacy config file if env vars are missing
	if bridgeIP == "" || apiKey == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			legacyPath := filepath.Join(home, ".hue-config.json")
			if data, err := os.ReadFile(legacyPath); err == nil {
				var legacyConfig struct {
					BridgeIP string `json:"bridge_ip"`
					APIKey   string `json:"api_key"`
				}
				if err := json.Unmarshal(data, &legacyConfig); err == nil {
					if bridgeIP == "" {
						bridgeIP = legacyConfig.BridgeIP
					}
					if apiKey == "" {
						apiKey = legacyConfig.APIKey
					}
				}
			}
		}
	}

	if bridgeIP == "" || apiKey == "" {
		return nil, fmt.Errorf("configuration not found. Set HUE_BRIDGE_IP and HUE_API_KEY environment variables, or run 'hue-control setup'")
	}

	return &Config{
		BridgeIP: bridgeIP,
		APIKey:   apiKey,
	}, nil
}

func saveConfig(config *Config) error {
	// Check for existing .env to append/update, or create new
	envPath := ".env"

	// Simple .env writing (overwrites logic for simplicity in this tailored tool)
	content := fmt.Sprintf("HUE_BRIDGE_IP=%s\nHUE_API_KEY=%s\n", config.BridgeIP, config.APIKey)
	return os.WriteFile(envPath, []byte(content), 0600)
}

// getHTTPClient returns an HTTP client configured for Hue Bridge communication
func getHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Hue Bridge uses self-signed certs
			},
		},
	}
}

func runSetup() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Hue Bridge IP address: ")
	bridgeIP, _ := reader.ReadString('\n')
	bridgeIP = strings.TrimSpace(bridgeIP)

	if bridgeIP == "" {
		fmt.Println("Error: Bridge IP is required")
		os.Exit(1)
	}

	fmt.Println("\nPress the button on your Hue Bridge, then press Enter here...")
	reader.ReadString('\n')

	// Create user/API key
	apiKey, err := createUser(bridgeIP)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	config := &Config{
		BridgeIP: bridgeIP,
		APIKey:   apiKey,
	}

	if err := saveConfig(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSuccess! Configuration saved to .env\n")
	fmt.Println("You can now use 'hue-control list' to see your rooms.")
}

func createUser(bridgeIP string) (string, error) {
	client := getHTTPClient()
	url := fmt.Sprintf("https://%s/api", bridgeIP)

	body := map[string]string{
		"devicetype": "hue-control#cli",
	}
	jsonBody, _ := json.Marshal(body)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to bridge: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result []map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("invalid response from bridge")
	}

	if len(result) == 0 {
		return "", fmt.Errorf("empty response from bridge")
	}

	if errInfo, ok := result[0]["error"]; ok {
		errMap := errInfo.(map[string]interface{})
		return "", fmt.Errorf("%v", errMap["description"])
	}

	if successInfo, ok := result[0]["success"]; ok {
		successMap := successInfo.(map[string]interface{})
		if username, ok := successMap["username"]; ok {
			return username.(string), nil
		}
	}

	return "", fmt.Errorf("unexpected response from bridge")
}

func runList() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	groups, err := getGroups(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available Rooms/Groups:")
	fmt.Println("------------------------")
	for id, group := range groups {
		status := "off"
		if group.Action.On {
			brightness := int(float64(group.Action.Bri) / 254.0 * 100)
			status = fmt.Sprintf("on (%d%%)", brightness)
		}
		fmt.Printf("  [%s] %s (%s) - %d lights - %s\n", id, group.Name, group.Type, len(group.Lights), status)
	}
}

func getGroups(config *Config) (map[string]Group, error) {
	client := getHTTPClient()
	url := fmt.Sprintf("https://%s/api/%s/groups", config.BridgeIP, config.APIKey)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bridge: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var groups map[string]Group
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("invalid response: %v", err)
	}

	return groups, nil
}

func runSet() {
	setCmd := flag.NewFlagSet("set", flag.ExitOnError)
	room := setCmd.String("room", "all", "Room name to control")
	brightness := setCmd.Int("brightness", 100, "Brightness percentage (0-100)")
	hueVal := setCmd.Int("hue", -1, "Hue value (0-65535)")
	satVal := setCmd.Int("sat", -1, "Saturation value (0-254)")
	colorName := setCmd.String("color", "", "Color preset name")
	setCmd.Parse(os.Args[2:])

	if *brightness < 0 || *brightness > 100 {
		fmt.Println("Error: Brightness must be between 0 and 100")
		os.Exit(1)
	}

	// Resolve color preset
	var finalHue, finalSat int = -1, -1
	if *colorName != "" {
		preset, ok := ColorPresets[strings.ToLower(*colorName)]
		if !ok {
			fmt.Printf("Error: Unknown color '%s'. Available: red, orange, yellow, green, cyan, blue, purple, pink, warm, cool, white\n", *colorName)
			os.Exit(1)
		}
		finalHue = preset[0]
		finalSat = preset[1]
	}

	// Override with explicit hue/sat if provided
	if *hueVal >= 0 {
		if *hueVal > 65535 {
			fmt.Println("Error: Hue must be between 0 and 65535")
			os.Exit(1)
		}
		finalHue = *hueVal
	}
	if *satVal >= 0 {
		if *satVal > 254 {
			fmt.Println("Error: Saturation must be between 0 and 254")
			os.Exit(1)
		}
		finalSat = *satVal
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Convert percentage to Hue brightness (1-254)
	hueBrightness := int(float64(*brightness) / 100.0 * 254)
	if hueBrightness < 1 && *brightness > 0 {
		hueBrightness = 1
	}

	if strings.ToLower(*room) == "all" {
		err = setAllLights(config, true, hueBrightness, finalHue, finalSat)
	} else {
		err = setRoomState(config, *room, hueBrightness, finalHue, finalSat)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Build output message
	msg := fmt.Sprintf("Set %s to %d%% brightness", *room, *brightness)
	if finalHue >= 0 || finalSat >= 0 {
		if *colorName != "" {
			msg += fmt.Sprintf(" with color '%s'", *colorName)
		} else {
			msg += fmt.Sprintf(" with hue=%d sat=%d", finalHue, finalSat)
		}
	}
	fmt.Println(msg)
}

func runOn() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := setAllLights(config, true, 254, -1, -1); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All lights turned on")
}

func runOff() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := setAllLights(config, false, 0, -1, -1); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All lights turned off")
}

func setAllLights(config *Config, on bool, brightness int, hue int, sat int) error {
	groups, err := getGroups(config)
	if err != nil {
		return err
	}

	// Find group 0 (all lights) or iterate through all groups
	client := getHTTPClient()

	// Try to use the special "0" group which represents all lights
	url := fmt.Sprintf("https://%s/api/%s/groups/0/action", config.BridgeIP, config.APIKey)

	state := map[string]interface{}{
		"on": on,
	}
	if on && brightness > 0 {
		state["bri"] = brightness
	}
	if hue >= 0 {
		state["hue"] = hue
	}
	if sat >= 0 {
		state["sat"] = sat
	}

	jsonBody, _ := json.Marshal(state)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// Fall back to setting each group individually
		for id := range groups {
			url := fmt.Sprintf("https://%s/api/%s/groups/%s/action", config.BridgeIP, config.APIKey, id)
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
		return nil
	}
	defer resp.Body.Close()

	return nil
}

func setRoomState(config *Config, roomName string, brightness int, hue int, sat int) error {
	groups, err := getGroups(config)
	if err != nil {
		return err
	}

	// Find the group by name
	var groupID string
	for id, group := range groups {
		if strings.EqualFold(group.Name, roomName) {
			groupID = id
			break
		}
	}

	if groupID == "" {
		return fmt.Errorf("room '%s' not found. Use 'hue-control list' to see available rooms", roomName)
	}

	client := getHTTPClient()
	url := fmt.Sprintf("https://%s/api/%s/groups/%s/action", config.BridgeIP, config.APIKey, groupID)

	state := map[string]interface{}{
		"on":  true,
		"bri": brightness,
	}
	if hue >= 0 {
		state["hue"] = hue
	}
	if sat >= 0 {
		state["sat"] = sat
	}

	jsonBody, _ := json.Marshal(state)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to bridge: %v", err)
	}
	defer resp.Body.Close()

	return nil
}
