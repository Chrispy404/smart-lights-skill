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
	"time"
)

// WeatherCondition maps weather codes to color names
var WeatherConditions = map[string]string{
	// wttr.in WWO codes -> color names (matching hue-control presets)
	// Clear/Sunny
	"113": "warm", // Sunny
	"116": "warm", // Partly Cloudy

	// Cloudy
	"119": "cool", // Cloudy
	"122": "cool", // Overcast
	"143": "cool", // Mist
	"248": "cool", // Fog
	"260": "cool", // Freezing Fog

	// Rain
	"176": "blue", // Patchy rain
	"263": "blue", // Patchy light drizzle
	"266": "blue", // Light drizzle
	"293": "blue", // Patchy light rain
	"296": "blue", // Light rain
	"299": "blue", // Moderate rain at times
	"302": "blue", // Moderate rain
	"305": "blue", // Heavy rain at times
	"308": "blue", // Heavy rain
	"311": "cyan", // Light freezing rain
	"314": "cyan", // Moderate or heavy freezing rain
	"353": "blue", // Light rain shower
	"356": "blue", // Moderate or heavy rain shower
	"359": "blue", // Torrential rain shower

	// Snow
	"179": "white", // Patchy snow
	"182": "cyan",  // Patchy sleet
	"185": "cyan",  // Patchy freezing drizzle
	"227": "white", // Blowing snow
	"230": "white", // Blizzard
	"317": "cyan",  // Light sleet
	"320": "cyan",  // Moderate or heavy sleet
	"323": "white", // Patchy light snow
	"326": "white", // Light snow
	"329": "white", // Patchy moderate snow
	"332": "white", // Moderate snow
	"335": "white", // Patchy heavy snow
	"338": "white", // Heavy snow
	"350": "cyan",  // Ice pellets
	"362": "cyan",  // Light sleet showers
	"365": "cyan",  // Moderate or heavy sleet showers
	"368": "white", // Light snow showers
	"371": "white", // Moderate or heavy snow showers
	"374": "cyan",  // Light showers of ice pellets
	"377": "cyan",  // Moderate or heavy showers of ice pellets

	// Thunderstorm
	"200": "purple", // Thundery outbreaks
	"386": "purple", // Patchy light rain with thunder
	"389": "purple", // Moderate or heavy rain with thunder
	"392": "purple", // Patchy light snow with thunder
	"395": "purple", // Moderate or heavy snow with thunder
}

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

	// Determine color from weather code
	color, ok := WeatherConditions[weatherCode]
	if !ok {
		color = "warm" // Default to warm if unknown
	}

	fmt.Printf("📍 Location: %s\n", locationName)
	fmt.Printf("🌡️  Temperature: %s°C (feels like %s°C)\n", current.TempC, current.FeelsLikeC)
	fmt.Printf("☁️  Condition: %s (code: %s)\n", weatherDesc, weatherCode)
	fmt.Printf("💡 Setting lights to: %s at %d%% brightness\n", color, *brightness)

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
