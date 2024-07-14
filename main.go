// Traveltime is a CLI tool that returns the travel time between two locations.
// It also returns the traffic induced delay, so you can avoid long waits in traffic.
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"text/template"
	"time"

	"googlemaps.github.io/maps"
)

var (
	apiEnv          = "GOOGLE_API_KEY"
	workEnv         = "TRAVEL_WORK_COORD"
	homeEnv         = "TRAVEL_HOME_COORD"
	formatOutputEnv = "TRAVEL_FORMAT_OUTPUT"
	defaultFormat   = `{{ .Origin.Name }}: {{ .WithTraffic }} {{ .Deviation.Absolute }}min`
)

func main() {
	apiKey := os.Getenv(apiEnv)
	if apiKey == "" {
		log.Fatalf("missing api key, use %q to provide key.\n", apiEnv)
	}
	workArg := os.Getenv(workEnv)
	if workArg == "" {
		log.Fatalf("missing work coordinate, use %q to provide key.\n", workEnv)
	}
	homeArg := os.Getenv(homeEnv)
	if homeArg == "" {
		log.Fatalf("missing home coordinate, use %q to provide key.\n", homeEnv)
	}
	format := defaultFormat
	if customFormat := os.Getenv(formatOutputEnv); customFormat != "" {
		format = customFormat
	}

	outTemplate, err := template.New("output").Parse(format)
	if err != nil {
		log.Fatalf("invalid format %q: %e", defaultFormat, err)
	}
	work, err := parseLatLngName(workArg)
	if err != nil {
		log.Fatal(err)
	}
	home, err := parseLatLngName(homeArg)
	if err != nil {
		log.Fatal(err)
	}

	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	// Traveltime needs your current position to calculate which of the given locations is the origin.
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	locationResult, err := client.Geolocate(ctx, &maps.GeolocationRequest{ConsiderIP: true})
	if err != nil {
		log.Fatal("failed to fetch geolocation: ", err)
	}
	// Use the current position to calculate the origin.
	origin, destination := findDirection(work, home, locationResult.Location)
	// Call an upstream API for the optimal and actual travel duration.
	// For now use Google's Distance Matrix API.
	distanceResult, err := client.DistanceMatrix(ctx, &maps.DistanceMatrixRequest{
		Origins:       []string{origin.LatLng.String()},
		Destinations:  []string{destination.LatLng.String()},
		Mode:          maps.TravelModeDriving,
		DepartureTime: fmt.Sprintf("%d", time.Now().Unix()),
		TrafficModel:  maps.TrafficModelBestGuess,
	})
	if err != nil {
		log.Fatal("failed to fetch distance matrix: ", err)
	}
	// Traveltime is designed to return the optimal travel duration, the time you would travel without traffic.
	// It also returns the actual travel duration, the time you should plan considering the current traffic situation.
	//
	// Calculate those information from the API response.
	result := &TravelResult{
		Origin:      origin,
		Destination: destination,
		WithTraffic: int(getWithTrafficDuration(distanceResult).Minutes()),
		NoTraffic:   int(getNoTrafficeDuration(distanceResult).Minutes()),
		Deviation:   newDeviation(distanceResult),
	}

	if err := outTemplate.Execute(os.Stdout, result); err != nil {
		log.Fatal("failed to execute template: ", err)
	}
}

func getWithTrafficDuration(distanceResult *maps.DistanceMatrixResponse) time.Duration {
	return distanceResult.Rows[0].Elements[0].DurationInTraffic
}

func getNoTrafficeDuration(distanceResult *maps.DistanceMatrixResponse) time.Duration {
	return distanceResult.Rows[0].Elements[0].Duration
}

func getAbsoluteDeviation(distanceResult *maps.DistanceMatrixResponse) float64 {
	withTraffic := distanceResult.Rows[0].Elements[0].DurationInTraffic.Minutes()
	noTraffic := distanceResult.Rows[0].Elements[0].Duration.Minutes()
	return (withTraffic - noTraffic)
}

func getRelativeDeviation(distanceResult *maps.DistanceMatrixResponse) float64 {
	withTraffic := distanceResult.Rows[0].Elements[0].DurationInTraffic.Seconds()
	noTraffic := distanceResult.Rows[0].Elements[0].Duration.Seconds()
	return ((100 / noTraffic * withTraffic) - 100)
}

// TravelResult holds all informations about a travel.
// It contains different representations of the travel time and the deviation.
// All fields can be accessed by the output template.
type TravelResult struct {
	Origin, Destination LatLngName
	// WithTraffic is the calculated travel time under consideration of traffic induced delay.
	WithTraffic int
	// NoTraffic is the optimal travel time without any delay by traffic.
	NoTraffic int
	// Deviation contains the difference between NoTraffic and WithTraffic in different formats.
	Deviation Deviation
}

// Deviation contains different versions of the delay induced by traffic on the travel.
type Deviation struct {
	// Relative is the deviation in percent.
	Relative string
	// Absolute is the deviation in minutes.
	Absolute string
}

func newDeviation(distanceResult *maps.DistanceMatrixResponse) Deviation {
	return Deviation{
		Relative: fmt.Sprintf("%+d", int(getRelativeDeviation(distanceResult))),
		Absolute: fmt.Sprintf("%+d", int(getAbsoluteDeviation(distanceResult))),
	}
}

// findDirection calculates which coordinate is less far away from your current location.
// Based on this information in which direction you need to travel.
// Your origin is the nearest point to your current location.
func findDirection(pointA, pointB LatLngName, location maps.LatLng) (origin, destination LatLngName) {
	distance1 := calculateDistance(pointA.LatLng, location)
	distance2 := calculateDistance(pointB.LatLng, location)
	if distance1 < distance2 {
		return pointA, pointB
	}
	return pointB, pointA
}

func calculateDistance(point1, point2 maps.LatLng) float64 {
	return math.Sqrt(math.Pow(point2.Lat-point1.Lat, 2) + math.Pow(point2.Lng-point1.Lng, 2))
}

// LatLngName extends the googlemaps.github.io/maps.LatLng struct with a name.
//
// See `go doc googlemaps.github.io/maps.LatLng` for more information.
type LatLngName struct {
	maps.LatLng
	Name string
}

func parseLatLngName(location string) (LatLngName, error) {
	count := strings.Count(location, ",")
	if count != 2 {
		return LatLngName{}, fmt.Errorf("invalid format, must contain 2 ',', got %d", count)
	}
	name, latLng, ok := strings.Cut(location, ",")
	if !ok {
		return LatLngName{}, fmt.Errorf("failed to parse name")
	}
	var result LatLngName
	var err error
	result.LatLng, err = maps.ParseLatLng(latLng)
	if err != nil {
		return LatLngName{}, err
	}
	result.Name = name
	return result, nil
}
