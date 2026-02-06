package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// WttrResponse represents the wttr.in JSON response
type WttrResponse struct {
	CurrentCondition []struct {
		WeatherCode string `json:"weatherCode"`
		WeatherDesc []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
		TempC      string `json:"temp_C"`
		FeelsLikeC string `json:"FeelsLikeC"`
	} `json:"current_condition"`
	NearestArea []struct {
		AreaName []struct {
			Value string `json:"value"`
		} `json:"areaName"`
		Country []struct {
			Value string `json:"value"`
		} `json:"country"`
	} `json:"nearest_area"`
}

func main() {
	location := flag.String("location", "", "Location for weather (default: auto-detect)")
	room := flag.String("room", "all", "Room to control")
	brightness := flag.Int("brightness", 80, "Brightness percentage (0-100)")
	dryRun := flag.Bool("dry-run", false, "Show what would be done without executing")
	flag.Parse()

	// Fetch weather
	weather, err := getWeather(*location)
	if err != nil {
		fmt.Printf("Error fetching weather: %v\n", err)
		os.Exit(1)
	}

	if len(weather.CurrentCondition) == 0 {
		fmt.Println("Error: No weather data received")
		os.Exit(1)
	}

	current := weather.CurrentCondition[0]
	weatherCode := current.WeatherCode
	weatherDesc := "Unknown"
	if len(current.WeatherDesc) > 0 {
		weatherDesc = current.WeatherDesc[0].Value
	}

	locationName := "Unknown"
	if len(weather.NearestArea) > 0 && len(weather.NearestArea[0].AreaName) > 0 {
		locationName = weather.NearestArea[0].AreaName[0].Value
	}

	// Determine base color from weather code
	baseColor, ok := WeatherConditions[weatherCode]
	if !ok {
		baseColor = "warm" // Default to warm if unknown
	}

	// Adjust color based on temperature
	temp, _ := strconv.Atoi(current.TempC)
	color := adjustColorByTemperature(baseColor, temp)

	fmt.Printf("📍 Location: %s\n", locationName)
	fmt.Printf("🌡️  Temperature: %s°C (feels like %s°C)\n", current.TempC, current.FeelsLikeC)
	fmt.Printf("☁️  Condition: %s (code: %s)\n", weatherDesc, weatherCode)
	fmt.Printf("💡 Setting lights to: %s at %d%% brightness", color, *brightness)
	if color != baseColor {
		fmt.Printf(" (adjusted from %s due to temperature)\n", baseColor)
	} else {
		fmt.Println()
	}

	if *dryRun {
		fmt.Println("\n[Dry run - no changes made]")
		return
	}

	// Find hue-control binary
	hueControlPath, err := findHueControl()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Execute hue-control
	args := []string{"set", "--color", color, "--brightness", fmt.Sprintf("%d", *brightness)}
	if *room != "all" {
		args = append(args, "--room", *room)
	}

	cmd := exec.Command(hueControlPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error setting lights: %v\n", err)
		os.Exit(1)
	}
}

func getWeather(location string) (*WttrResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := "https://wttr.in/?format=j1"
	if location != "" {
		url = fmt.Sprintf("https://wttr.in/%s?format=j1", location)
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var weather WttrResponse
	if err := json.Unmarshal(body, &weather); err != nil {
		return nil, fmt.Errorf("failed to parse weather data: %v", err)
	}

	return &weather, nil
}

func findHueControl() (string, error) {
	// Try relative path first (same directory structure)
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		// Check sibling directory
		siblingPath := filepath.Join(dir, "..", "hue-control", "hue-control")
		if _, err := os.Stat(siblingPath); err == nil {
			return siblingPath, nil
		}
		// Check same directory
		sameDirPath := filepath.Join(dir, "hue-control")
		if _, err := os.Stat(sameDirPath); err == nil {
			return sameDirPath, nil
		}
	}

	// Try PATH
	path, err := exec.LookPath("hue-control")
	if err == nil {
		return path, nil
	}

	// Try common locations
	commonPaths := []string{
		"./hue-control",
		"../hue-control/hue-control",
		"./scripts/hue-control/hue-control",
	}
	for _, p := range commonPaths {
		if abs, err := filepath.Abs(p); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
	}

	return "", fmt.Errorf("hue-control binary not found. Build it first with ./scripts/build.sh")
}
