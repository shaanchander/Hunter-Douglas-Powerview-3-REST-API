(currently assumes the blind has a shade for initial implementation)

TODO
- Account for slight percentage difference (doesn't add to 100 due to compacted blind taking up space)
- Handle blinds with no shades
- implement caching for shade types?
- query shades by int id?
- make container to host api
- improve error handling
- improve response from api

Create a local `config.yaml` by copying the sample file and filling in your values:

```bash
cp config.sample.yaml config.yaml
```

Then edit `config.yaml` with your PowerView host and API port.

Start the API with:

```bash
go run ./cmd/api
```

The server reads `config.yaml` from the repository root.

```yaml
POWERVIEW_HOST: http://192.168.1.50
API_PORT: "8080"
```

Then start the API with:

```bash
go run ./cmd/api
```

## Run With Docker

Build the image:

```bash
docker build -t powerview-api .
```

Run it and mount your local config file:

```bash
docker run --rm \
	-p 8080:8080 \
	-v "$(pwd)/config.yaml:/app/config.yaml:ro" \
	powerview-api
```

If you changed `API_PORT` in `config.yaml`, update the `-p host:container` mapping to match.

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