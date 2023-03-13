# traveltime
Cli tool to fetch traffic data between given locations.
Traveltime will calculate which of the two configured locations is your commute destination, based on your current location.
It fetches your current location and selects the most distant location as destination.

## Setup
It uses Googles [Geolocation API](https://developers.google.com/maps/documentation/geolocation/overview) and [Distance Matrix API](https://developers.google.com/maps/documentation/distance-matrix).
You need a google API key to use the services.
Provide one using the environment variable `GOOGLE_API_KEY`.
The locations are also configured by environment variables `TRAVEL_WORK_COORD` and `TRAVEL_HOME_COORD`.
They are formatted like this: `<name>,<lat>,<long>`.
