# telegraf-hoymiles-wifi
Telegraf Plugin for Hoymiles Wifi Inverters

## Running as a `telegraf` plugin

- Build the binary, i.e. via `go build -o hoymiles ../cmd/hoymiles`
- Add a configuration for the plugin (take a look at [`sample.conf`](plugins/inputs/hoymiles_wifi/sample.conf))
- Add to your `telegraf.conf`
  ```toml
  [[inputs.execd]]
    command = ["hoymiles -config my-config.conf"]
    signal = "none"
  ```
- Run telegraf

## Bundling in `gokrazy`

### `config.json`

```jsonc
{
    "Packages": [
        // ...
        "github.com/liwde/telegraf-hoymiles-wifi/cmd/hoymiles",
        "github.com/influxdata/telegraf/cmd/telegraf"
    ],
    "PackageConfig": {
        // ...
        "github.com/liwde/telegraf-hoymiles-wifi/cmd/hoymiles": {
            "DontStart": true
        },
        "github.com/influxdata/telegraf/cmd/telegraf": {
            "ExtraFilePaths": {
                "/etc/telegraf/telegraf.conf": "./telegraf.toml",
                "/etc/telegraf/hoymiles.conf": "./hoymiles.toml",
            },
            "Environment": [
                "INFLUX_TOKEN=<your-token>"
            ],
            "WaitForClock": true
        }
    },
    // ...
}
```

### `telgraf.toml`

```toml
[[inputs.execd]]
  command = ["/user/hoymiles", "-config", "/etc/telegraf/hoymiles.conf"]
  signal = "none"

[[outputs.influxdb_v2]]
  urls = ["<your-influx-url>"]
  token = "${INFLUX_TOKEN}"
  organization = "<your-organization>"
  bucket = "<your-bucket>"
```

### `hoymiles.toml`

```toml
[[inputs.hoymiles_wifi]]
  hostname = "<your-hoymiles-hostname>"
```

## Thanks

Based on the fabulous work of https://github.com/BLun78/hoymiles_wifi
