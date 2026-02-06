# Philips Hue & Weather Skill 💡🌦️

An agentic skill for controlling Philips Hue lights via CLI, with intelligent weather-based automation.

## Features

- **💡 Light Control**: Turn on/off, set brightness, and control specific rooms.
- **🎨 Color Control**: Set colors using presets (`blue`, `warm`, `cool`) or precise hue/saturation.
- **🌦️ Weather Mode**: Automatically sets light "mood" based on your local weather and temperature (e.g., Freezing Sunny -> Cool White, Rainy -> Blue).
- **🔒 Secure**: Uses `.env` for secure credential storage.
- **🤖 Agent Ready**: Includes workflows for AI agents.

## Quick Start

### 1. Build
```bash
./scripts/build.sh
```

### 2. Setup
Connect to your Hue Bridge:
```bash
./scripts/hue-control/hue-control setup
```
(Press the button on your Bridge when prompted)

### 3. Usage
**Control Lights:**
```bash
# Turn all on
./scripts/hue-control/hue-control on

# Set warm mood in Living Room
./scripts/hue-control/hue-control set --room "Living Room" --color warm --brightness 80
```

**Check Weather Mode:**
```bash
# Set lights based on current location's weather
./scripts/weather-lights/weather-lights
```

## Documentation

For full details, see [SKILL.md](SKILL.md).
