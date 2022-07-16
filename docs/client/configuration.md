# 🛠️ Configuration

## 🚩 Flags and environment variables

Available flags:

|        Flag         | Short |           Description           | Default value |
|:-------------------:|:-----:|:-------------------------------:|:-------------:|
|     `-settings`     |   /   |   Path to the settings file.    |               |
| `-connector-folder` |   /   |  Path to the connector folder.  |               |
|   `-ocpp-config`    |   /   | Path to the OCPP configuration. |               |
|       `-auth`       |   /   | Path to the authorization file. |               |
|      `-debug`       | `--d` |           Debug mode            |     false     |
|       `-api`        | `--a` |         Expose the API          |     false     |
|   `-api-address`    |   /   |           API address           |  "localhost"  |
|     `-api-port`     |   /   |            API port             |     4269      |

Environment variables are created automatically thanks to [Viper](https://github.com/spf13/viper) and are prefixed
with `CHARGEPI`. Only the settings (not the ocpp configuration or connectors) are bound to the env
variables. Debug mode and API settings flags are also bound to the environment variables.

Example environment variable: `CHARGEPI_CHARGEPOINT_INFO_ID`.

## 🛠 Configuration files

There are three **required** configuration files:

1. [`settings`](../../configs/settings.json)
2. [`configuration`](../../configs/configuration.json)
3. [`connector`](../../configs/connectors/connector-1.json)

The settings files are supported in `YAML`, `JSON` or `TOML` format.

### The `settings` file

The `settings` file contains basic information about the charge point and provides connectivity details:

- Charge Point ID,
- central system URI and OCPP protocol version,
- OCPP information (vendor, model, firmware, etc.),
- logging settings,
- TLS settings,
- default max charging time,
- hardware settings for LCD, RFID/NFC reader and LEDs.

The table represents attributes, their values and descriptions that require more attention and might not be
self-explanatory. Some attributes can have multiple possible values, if any are empty, they will be treated as disabled
or might not work properly.

|        Attribute        |                                  Description                                  |                         Possible values                          | 
|:-----------------------:|:-----------------------------------------------------------------------------:|:----------------------------------------------------------------:|
|           id            |      ID of the charging point. Must be registered in the Central System       |                        Default:"ChargePi"                        |
|     protocolVersion     |                         Version of the OCPP protocol.                         |                          "1.6", "2.0.1"                          |
|        serverUri        |             URI of the Central System with the port and endpoint.             | Default: "172.0.1.121:8080/steve/websocket/CentralSystemService" |
|  info: maxChargingTime  |          Max charging time allowed on the Charging point in minutes.          |                           Default:180                            |
| rfidReader: readerModel |                          RFID/NFC reader model used.                          |                           "PN532", ""                            | 
|   ledIndicator: type    |                          Type of the led indicator.                           |                           "WS281x", ""                           |
|   hardware: minPower    | Minimum power draw needed to continue charging, if Power meter is configured. |                            Default:20                            |

Example settings:

```json
{
  "chargePoint": {
    "info": {
      "id": "ChargePi",
      "protocolVersion": "1.6",
      "serverUri": "example.com",
      "basicAuthUser": "",
      "basicAuthPass": "",
      "maxChargingTime": 5,
      "ocpp": {
        "vendor": "UL FE",
        "model": "ChargePi"
      }
    },
    "logging": {
      "type": [
        "remote",
        "file"
      ],
      "format": "gelf",
      "host": "logging.example.com",
      "port": 12201
    },
    "tls": {
      "isEnabled": false,
      "CACertificatePath": "/usr/share/certs/rootCA.crt",
      "clientCertificatePath": "/usr/share/certs/charge-point.crt",
      "clientKeyPath": "/usr/share/certs/charge-point.key"
    },
    "hardware": {
      "lcd": {
        "isSupported": true,
        "driver": "hd44780",
        "i2cAddress": "0x27",
        "i2cBus": 1,
        "language": "en"
      },
      "tagReader": {
        "isSupported": true,
        "readerModel": "PN532",
        "device": "/dev/ttyS0",
        "resetPin": 19
      },
      "ledIndicator": {
        "enabled": true,
        "type": "WS281x",
        "dataPin": 18,
        "indicateCardRead": true,
        "invert": false
      },
      "powerMeters": {
        "minPower": 20,
        "retries": 3
      }
    }
  }
}
```

### 🔌 The `connector` file(s) - EVSEs and connectors

EVSE and connector settings files can be found in the `connectors` folder. To add and configure the connector, simply
add a new file that contains the structure, defined in [attributes](#attributes) and modify it to your specs. The client
will scan the folder at boot and configure the connectors from the files if all the settings have valid values.

Note: A Charge point can have multiple EVSEs, each oh which can have multiple connectors, but only one connector of the
EVSE can charge at a time.

#### Attributes

`Connector` object contains a connector type and an ID of the connector, which must start with 1 and increment by one.
The status attribute changes according to the OCPP specification. The `session` represents a Charging session and is
used to restore the connector's last state when (re)starting the client.

The `relay` and `powerMeter` objects are configurable to specific GPIO pins and SPI bus. The `inverseLogic` attribute in
the relay object indicates the logic of the relay. If `inverseLogic` is set to _true_, the relay will use negative
logic. The `powerMeter` also contains some attributes for measurement calibration.

The table represents attributes, their values and descriptions that require more attention and might not be
self-explanatory. Some attributes can have multiple possible values, if any are empty, they will be treated as disabled
or might not work properly.

|            Attribute             |                           Description                            |                Possible values                 | 
|:--------------------------------:|:----------------------------------------------------------------:|:----------------------------------------------:|
|              evseId              |                          ID of the EVSE                          |                       /                        |
|               type               |            A type of the connector used in the build.            | Refer to OCPP documentation. Default: "Schuko" |
|       relay: inverseLogic        |         Uses negative logic for operating with the relay         |                     false                      | 
|     powerMeter: shuntOffset      | Value of the shunt resistor used in the build to measure power.  |                 Default: 0.01                  | 
| powerMeter: voltageDividerOffset | Value of the voltage divider used in the build to measure power. |                  Default:1333                  |

Example connector:

```json
{
  "evseId": 1,
  "connectorId": 1,
  "type": "Schuko",
  "status": "Available",
  "session": {
    "isActive": false,
    "transactionId": "",
    "tagId": "",
    "started": "",
    "consumption": []
  },
  "evcc": {
    "relayPin": 26,
    "inverseLogic": false
  },
  "powerMeter": {
    "enabled": false,
    "type": "CS5460A",
    "powerMeterPin": 25,
    "spiBus": 0,
    "powerUnits": "kWh",
    "consumption": 0.0,
    "shuntOffset": 0.055,
    "voltageDividerOffset": 1333
  }
}
```