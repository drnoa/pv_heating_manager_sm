# Heating Manager
The Heating Manager is a Go-based program designed for monitoring and controlling a heating system using Solar-Manager. It continuously checks the temperature and activates the heating if needed to prevent exceeding a predefined temperature threshold. Additionally, it conducts weekly checks to ensure the system's proper functionality.

## Features

- **Temperature Monitoring**: Monitors the temperature via a Shelly device and logs the status.
- **Automatic Heating Control**: Turns on the heating when the set temperature threshold is exceeded.
- **Weekly System Check**: Performs automatic weekly checks to ensure the system's operability.

## Configuration

The program reads its configuration from a file named `config.json`, which should be placed in the same directory as the program. The configuration file should include the following settings:

```json
{
    "solarManagerURL": "https://cloud.solar-manager.ch/v1/stream/sensor/{solarmanagerid}",
    "solarManagerSensorID": "solarmanagertempsensor_id",
    "temperatureThreshold": 55,
    "temperatureTurnOff": 60,
    "checkInterval": 5, 
    "weeklyCheckInterval": 168,
    "username": "username",
    "password": "password"
}
```

## Installation
Ensure Go is installed on your system.
Clone the repository or download the source files.
Run go build in the project directory to create the executable file.
## Usage
After configuring config.json appropriately and compiling the program, you can start the Heating Manager by running the generated executable:

```bash
./heating_manager
```

The program will continue to run in the background, monitoring the temperature and controlling the heating as needed.

## License
This project is licensed under the GNU General Public License v3.0 - see the LICENSE file for details.