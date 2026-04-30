(currently assumes the blind has a shade for initial implementation)

TODO
- Account for slight percentage difference (doesn't add to 100 due to compacted blind taking up space)
- Handle blinds with no shades
- implement caching for shade types?
- query shades by int id?
- make container to host api
- improve error handling
- improve response from api

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

## API Notes: POST /v1/position

Current API contract uses `selectedShade` + `blindPct`, with `shadePct` only for shade+blind devices.

Request fields:
- `selectedShade` (string, required): BLE shade name.
- `blindPct` (int, required): Blind coverage percent in `[0..100]`.
- `shadePct` (int, conditional):
	- Required for shade type `9` (shade+blind).
	- Must be omitted for shade type `6` (blind-only).
- `velocity` (int, optional): When provided, must be in `[10..255]`.

Validation and normalization:
- Shade type `9`: `shadePct` and `blindPct` must both be in `[0..100]`, and `shadePct + blindPct <= 100`.
- Shade type `6`: `shadePct` must be omitted, and `blindPct` must be in `[0..100]`.
- The API derives protocol `gapPct` internally:
	- Type `9`: `gapPct = 100 - shadePct - blindPct`
	- Type `6`: `gapPct = 100 - blindPct` (with internal `shadePct = 0`)

Response fields:
- `sentHex`
- `shadePct` (returns `-1` sentinel for type `6` blind-only shades)
- `blindPct`
- `upstreamStatusCode`