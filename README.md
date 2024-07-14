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

## Customization
Traveltimes output format can be customized by passing a `TRAVEL_FORMAT_OUTPUT` environment variable.
It features the [text/template](https://pkg.go.dev/text/template) package and defaults to `{{ .Origin.Name }}: {{ .WithTraffic }} {{ .Deviation.Absolute }}min`.
You can access the fields of:
 * LatLngName
 ```Go
type LatLngName struct {
        maps.LatLng
        Name string
}
// LatLngName extends the googlemaps.github.io/maps.LatLng struct with a name.
//
// See `go doc googlemaps.github.io/maps.LatLng` for more information.
 ```
 * Deviation
 ```Go
type Deviation struct {
        // Relative is the deviation in percent.
        Relative string
        // Absolute is the deviation in minutes.
        Absolute string
}
// Deviation contains different versions of the delay induced by traffic on the travel.
 ```
 * TravelResult (accessible with the '.' operator from within the template)
```Go
type TravelResult struct {
        Origin, Destination LatLngName
        // WithTraffic is the calculated travel time under consideration of traffic induced delay.
        WithTraffic int
        // NoTraffic is the optimal travel time without any delay by traffic.
        NoTraffic int
        // Deviation contains the difference between NoTraffic and WithTraffic in different formats.
        Deviation Deviation
}
// TravelResult holds all informations about a travel. It contains different
// representations of the travel time and the deviation. All fields can be
// accessed by the output template.
```
