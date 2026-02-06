package main

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

// adjustColorByTemperature modifies the base color based on temperature
func adjustColorByTemperature(baseColor string, tempC int) string {
	// Temperature ranges influence color choice
	switch {
	case tempC < 0:
		// Freezing: shift to cooler tones
		if baseColor == "warm" || baseColor == "orange" {
			return "cool" // Sunny but freezing = cool white
		}
		if baseColor == "yellow" {
			return "cyan" // Warmer colors become cooler
		}
		return baseColor // Already cool colors stay

	case tempC >= 0 && tempC < 10:
		// Cold: slight cooling
		if baseColor == "warm" {
			return "white" // Sunny but cold = neutral white
		}
		return baseColor

	case tempC >= 30 && tempC < 38:
		// Hot: shift to warmer tones
		if baseColor == "cool" || baseColor == "white" {
			return "warm" // Cloudy but hot = warm
		}
		if baseColor == "warm" {
			return "orange" // Make it warmer
		}
		return baseColor

	case tempC >= 38:
		// Very hot: maximum warm/orange
		if baseColor == "cool" || baseColor == "white" || baseColor == "warm" {
			return "orange"
		}
		if baseColor == "yellow" {
			return "red" // Extreme heat
		}
		return baseColor

	default:
		// Moderate temps (10-30°C): use weather-based color as-is
		return baseColor
	}
}
