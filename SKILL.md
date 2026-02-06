---
name: Philips Hue Light Control
description: Control Philips Hue smart lights - set brightness, turn on/off rooms or all lights
---

# Philips Hue Light Control Skill

Control Philips Hue smart lights via the local Bridge API. Set brightness levels for specific rooms or all lights, and turn lights on/off.

## Prerequisites

- Philips Hue Bridge on your local network
- One-time setup to authenticate with the Bridge

## Setup

### 1. Build the CLI Tool

```bash
cd /path/to/lights-skill/scripts
./build.sh
```

### 2. Configure Bridge Connection

Run the setup command to generate a `.env` file:

```bash
./scripts/hue-control/hue-control setup
```

You will need to:
1. **Find your Bridge IP**: Go to [discovery.meethue.com](https://discovery.meethue.com) while on your home network. It will show your Bridge's "internalipaddress".
2. Enter that IP when prompted.
3. **Press the physical button** on your Hue Bridge when the script asks.
4. The tool will automatically generate and save your API key to `.env`.

## Usage

### List Available Rooms

```bash
./scripts/hue-control/hue-control list
```

### Set Brightness

Set brightness for all lights (defaults to 100%):
```bash
./scripts/hue-control/hue-control set
```

Set specific brightness percentage:
```bash
./scripts/hue-control/hue-control set --brightness 50
```

Set brightness for a specific room:
```bash
./scripts/hue-control/hue-control set --room "Living Room" --brightness 75
```

### Set Light Color

Use preset colors:
```bash
./scripts/hue-control/hue-control set --color blue
./scripts/hue-control/hue-control set --color warm --brightness 60
```

Available presets: `red`, `orange`, `yellow`, `green`, `cyan`, `blue`, `purple`, `pink`, `warm`, `cool`, `white`

Or use precise hue/saturation values:
```bash
./scripts/hue-control/hue-control set --hue 46920 --sat 254  # Blue
```

### Turn All Lights On/Off

```bash
./scripts/hue-control/hue-control on
./scripts/hue-control/hue-control off
```

### Weather-Based Lighting

Automatically set light colors based on current weather:
```bash
./scripts/weather-lights/weather-lights
```

Options:
- `--location "City"` - Specify location (default: auto-detect)
- `--room "Room"` - Target specific room
- `--brightness 80` - Set brightness (default: 80)
- `--dry-run` - Preview without changes

## Parameters

| Command | Parameter | Default | Description |
|---------|-----------|---------|-------------|
| `set` | `--room` | `all` | Room name to control, or "all" for all lights |
| `set` | `--brightness` | `100` | Brightness percentage (0-100) |

## Configuration

Configuration is stored in a `.env` file in the current directory, or can be provided via environment variables:

- `HUE_BRIDGE_IP`: IP address of the Hue Bridge
- `HUE_API_KEY`: Authenticated username/API key

See `.env.example` for the expected format.

> **Note**: For backward compatibility, the tool will also check `~/.hue-config.json` if environment variables are missing.
