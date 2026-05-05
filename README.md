# Hunter Douglas PowerView 3 REST API

A lightweight REST API that proxies and extends the Hunter Douglas PowerView 3 motorized blind hub. It provides endpoints to list registered shades, inspect shade details, and set shade/blind positions — translating user-friendly percentage values into the proprietary BLE protocol packets the hub expects.

## Run Locally

Start the API with flags:

```bash
go run ./cmd/api -H http://192.168.1.50 -P 8080
```

`-P` is optional and defaults to `8080`:

```bash
go run ./cmd/api -H http://192.168.1.50
```

## Run With Docker

Build the image:

```bash
docker build -t powerview-api .
```

Run it with environment variables:

```bash
docker run --rm \
	-p 8080:8080 \
	-e POWERVIEW_HOST=http://192.168.1.50 \
	-e API_PORT=8080 \
	powerview-api
```

If you set `API_PORT` to a non-default value, update the `-p host:container` mapping to match.

## API Endpoints

### GET /v1/gateway

Proxies the PowerView hub's `/gateway` endpoint. Returns raw gateway status (JSON).

**Response:** Raw JSON from the PowerView hub.

**Error:** `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/shades

Lists all shades registered in the PowerView home.

**Response:** Array of shade objects:

```json
[
  {
    "id": 1,
    "type": 9,
    "name": "Living Room",
    "ptName": "Living Room",
    "bleName": "HD-001A2B3C",
    "serialNumber": "ABC123",
    "signalStrength": -45,
    "roomId": 1,
    "batteryStatus": 3,
    "powerType": 1,
    "capabilities": 0,
    "positions": {
      "primary": 50.0,
      "secondary": 0.0,
      "tilt": 0.0,
      "velocity": 0
    }
  }
]
```

**Error:** `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/shades/:id

Gets a single shade by its integer ID.

**Path Parameters:**
- `id` (int, required): The shade's numeric ID.

**Response:** Single shade object (same schema as above).

**Errors:**
- `400 Bad Request` if `id` is not a valid integer.
- `404 Not Found` if no shade matches the given ID.
- `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/shades/:id/jog

Sends a jog command to a shade, causing it to briefly move (useful for identifying which physical shade corresponds to a given ID).

**Path Parameters:**
- `id` (int, required): The shade's numeric ID.

**Response:**

```json
{
  "shadeId": 1,
  "sentHex": "F7110A0103",
  "upstreamStatusCode": 200
}
```

**Errors:**
- `400 Bad Request` if `id` is not a valid integer.
- `404 Not Found` if no shade matches the given ID.
- `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/rooms

Lists all rooms registered in the PowerView home.

**Response:** Array of room objects:

```json
[
  {
    "id": 1,
    "ptName": "Living Room",
    "color": "#FF0000",
    "icon": "living-room",
    "type": 1,
    "shadeGroups": [
      {
        "id": 1,
        "ptName": "Group 1",
        "order": 0,
        "shadeIds": [1, 2]
      }
    ]
  }
]
```

**Error:** `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/rooms/:id

Gets a single room by its integer ID, including all shades in that room.

**Path Parameters:**
- `id` (int, required): The room's numeric ID.

**Response:** Room detail object with nested shades (firmware info stripped):

```json
{
  "id": 1,
  "name": "Living Room",
  "ptName": "Living Room",
  "color": "#FF0000",
  "icon": "living-room",
  "type": 1,
  "shadeGroups": [...],
  "shades": [...]
}
```

**Errors:**
- `400 Bad Request` if `id` is not a valid integer.
- `404 Not Found` if no room matches the given ID.
- `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/scenes

Lists all scenes registered in the PowerView home.

**Response:** Array of scene objects:

```json
[
  {
    "id": 108,
    "name": "Scene Name",
    "networkNumber": 1,
    "color": "#FF0000",
    "icon": "sun",
    "roomIds": [1, 2],
    "shadeIds": [1, 2, 3]
  }
]
```

**Error:** `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/scenes/:id

Gets a single scene by its integer ID. Returns the raw JSON response from the PowerView hub without processing.

**Path Parameters:**
- `id` (int, required): The scene's numeric ID.

**Response:** Raw JSON from the PowerView hub:

```json
{
  "id": 108,
  "name": "Scene Name",
  "networkNumber": 1,
  "color": "#FF0000",
  "icon": "sun",
  "roomIds": [1, 2],
  "shadeIds": [1, 2, 3]
}
```

**Errors:**
- `400 Bad Request` if `id` is not a valid integer.
- `404 Not Found` if no scene matches the given ID.
- `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/scenes/:id/trigger

Triggers a scene on the PowerView hub.

**Path Parameters:**
- `id` (int, required): The scene's numeric ID.

**Response:**

```json
{
  "sceneId": 108,
  "upstreamStatusCodes": [200, 200]
}
```

**Errors:**
- `400 Bad Request` if `id` is not a valid integer.
- `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/discover/ready

Checks whether the PowerView hub is ready to start a BLE discovery scan. Proxies the raw response from the PowerView hub's `/gateway/shades/discover/ready` endpoint. Use this before calling `/v1/discover` to avoid conflicts with an in-progress scan.

**Response:** Raw JSON from the PowerView hub:

```json
{
  "ready": true,
  "notReadyDevices": []
}
```

- `ready` — `true` if the hub is available for discovery, `false` if a scan is already in progress.
- `notReadyDevices` — array of device identifiers that are currently blocking discovery (empty when `ready` is `true`).

**Error:** `502 Bad Gateway` if the PowerView hub is unreachable.

---

### GET /v1/discover

Triggers a BLE scan on the PowerView hub to discover nearby shades (including unregistered ones). Proxies the raw response from the PowerView hub's `/gateway/shades/discover/` endpoint. This endpoint may take several seconds to complete while the hub performs the BLE scan.

Before starting the scan, this endpoint checks `/gateway/shades/discover/ready`. If the hub reports it is not ready (e.g., a scan is already in progress), it returns a `409 Conflict` instead of attempting the discovery.

**Response:** Raw JSON from the PowerView hub:

```json
{
  "scan": [
    {"bleName": "DUE:1C77", "filteredRssi": -56},
    {"bleName": "DUE:F354", "filteredRssi": -61},
    {"bleName": "DUE:51DB", "filteredRssi": -64}
  ],
  "excluded": []
}
```

- `scan` — array of discovered BLE devices with their filtered RSSI signal strength.
- `excluded` — (not exactly sure)

**Errors:**
- `409 Conflict` if the gateway is not ready for discovery (a scan is already in progress).
- `502 Bad Gateway` if the PowerView hub is unreachable.

---

### POST /v1/position

Sets the position of a shade/blind. Uses `id` to identify the shade, with `shadePct` and `blindPct` controlling the two layers.

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | int | Yes | Shade ID (from `/v1/shades`). |
| `blindPct` | int | Yes | Blind (blackout layer) coverage percent in `[0..100]`. |
| `shadePct` | int | Conditional | Shade (privacy layer) coverage percent in `[0..100]`. Required for shade type `9`; must be omitted for type `6`. |
| `velocity` | int | No | Motor speed in `[10..255]`. Omit for default speed. |

**Validation and normalization:**
- **Shade type `9` (shade+blind):** `shadePct` and `blindPct` must both be in `[0..100]`, and `shadePct + blindPct <= 100`. The API derives `gapPct = 100 - shadePct - blindPct`.
- **Shade type `6` (blind-only):** `shadePct` must be omitted, and `blindPct` must be in `[0..100]`. The API derives `gapPct = 100 - blindPct` internally.

**Response:**

```json
{
  "sentHex": "F7010109...",
  "shadePct": 40,
  "blindPct": 30,
  "upstreamStatusCode": 200
}
```

- `shadePct` returns `-1` sentinel for type `6` (blind-only) shades.

**Errors:**
- `400 Bad Request` for invalid body, missing fields, out-of-range values, or unsupported shade type.
- `502 Bad Gateway` if the PowerView hub is unreachable.

---

## TODO

- Account for slight percentage difference (doesn't add to 100 due to compacted blind taking up space)
- Handle blinds with no shades
- Implement caching for shade types
- Query shades by int ID
- Make container to host API
- Improve error handling
- Improve response from API
- add ability to jog/identify shade (done — see `GET /v1/shades/:id/jog`)