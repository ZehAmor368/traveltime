package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"googlemaps.github.io/maps"
)

var (
	apiEnv  string = "GOOGLE_API_KEY"
	workEnv string = "TRAVEL_WORK_COORD"
	homeEnv string = "TRAVEL_HOME_COORD"
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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	locationResult, err := client.Geolocate(ctx, &maps.GeolocationRequest{ConsiderIP: true})
	if err != nil {
		log.Fatal("failed to fetch geolocation: ", err)
	}
	origin, destination := findDirection(work, home, locationResult.Location)
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

	durationInTrafficSec := math.RoundToEven(distanceResult.Rows[0].Elements[0].DurationInTraffic.Seconds())
	durationInTrafficMin := math.RoundToEven(distanceResult.Rows[0].Elements[0].DurationInTraffic.Minutes())
	durationSec := math.RoundToEven(distanceResult.Rows[0].Elements[0].Duration.Seconds())
	deviation := (100 / durationSec * durationInTrafficSec) - 100

	fmt.Printf("%s: %dmin %+d%%\n", destination.Name, int(durationInTrafficMin), int(deviation))
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
