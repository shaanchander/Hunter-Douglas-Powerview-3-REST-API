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