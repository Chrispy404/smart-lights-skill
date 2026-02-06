---
description: Set brightness for a specific room
---

To set the brightness of a specific room (e.g. "Living Room") to a certain level:

1. List available rooms if names are unknown:
```bash
./scripts/hue-control/hue-control list
```

2. Run the set command with the room and brightness parameters:
// turbo
```bash
./scripts/hue-control/hue-control set --room "[Room Name]" --brightness [0-100]
```
