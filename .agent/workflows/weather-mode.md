---
description: Set light colors based on current weather
---

To set your lights to match the current weather:

// turbo
1. Run the weather-lights tool:
```bash
./scripts/weather-lights/weather-lights
```

Optional parameters:
- `--location "City Name"` - Specify location (default: auto-detect)
- `--room "Room Name"` - Target specific room (default: all)
- `--brightness 80` - Set brightness 0-100 (default: 80)
- `--dry-run` - Preview without changing lights

Example with options:
```bash
./scripts/weather-lights/weather-lights --location "Sydney" --brightness 60
```

Weather-to-color mapping:
- ☀️ Sunny/Clear → Warm orange
- ☁️ Cloudy/Overcast → Cool white
- 🌧️ Rain → Blue
- ❄️ Snow → White
- ⛈️ Thunderstorm → Purple
