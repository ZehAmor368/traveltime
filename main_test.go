package main

import (
	"bytes"
	"log"
	"testing"
	"text/template"
)

func TestTemplateDefaultFormat(t *testing.T) {
	expected := "home: 110 +10min"

	outTemplate, err := template.New("output").Parse(defaultFormat)
	if err != nil {
		log.Fatalf("invalid format %q: %s", defaultFormat, err)
	}
	result := &TravelResult{
		Origin: LatLngName{
			Name: "home",
		},
		Destination: LatLngName{
			Name: "work",
		},
		WithTraffic: 110,
		NoTraffic:   100,
		Deviation:   Deviation{Relative: "+10%", Absolute: "+10"},
	}

	var buf bytes.Buffer
	outTemplate.Execute(&buf, result)
	if expected != buf.String() {
		t.Fatalf("template returned unexpected format. got=%q want=%q", buf.String(), expected)
	}
}
