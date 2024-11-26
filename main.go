package main

import (
	"encoding/xml"
	"fmt"
	"log"
)

// TraFF represents the root element of the TraFF XML structure.
type TraFF struct {
	XMLName  xml.Name  `xml:"traff"`
	Messages []Message `xml:"message"`
}

// Message represents each traffic message in the TraFF XML.
type Message struct {
	XMLName  xml.Name `xml:"message"`
	Location Location `xml:"location"`
	Events   []Event  `xml:"events>event"`
}

// Location encapsulates the geographical details of the traffic event.
type Location struct {
	From         Point    `xml:"from"`
	To           Point    `xml:"to"`
	Polyline     Polyline `xml:"polyline"`
	JunctionRef  string   `xml:"junction_ref,attr,omitempty"`
	JunctionName string   `xml:"junction_name,attr,omitempty"`
}

// Point represents a geographical point with latitude and longitude.
type Point struct {
	Lat float64 `xml:"lat"`
	Lon float64 `xml:"lon"`
}

// Polyline represents a series of geographical coordinates.
type Polyline struct {
	Coordinates []Coordinate `xml:"coordinates>coordinate"`
}

// Coordinate represents a single geographical coordinate.
type Coordinate struct {
	Lon float64 `xml:"lon"`
	Lat float64 `xml:"lat"`
}

// Event represents a traffic event with its details.
type Event struct {
	Class       string `xml:"class,attr"`
	Type        string `xml:"type,attr"`
	Speed       string `xml:"speed,attr,omitempty"`
	Description string `xml:"description,omitempty"`
}

// Mapping from Autobahn API abnormalTrafficType to TraFF Class and Type.
var autobahnToTraFF = map[string]struct {
	Class string
	Type  string
}{
	"STATIONARY_TRAFFIC": {"CONGESTION", "CONGESTION_QUEUE"},
	"SLOW_TRAFFIC":       {"CONGESTION", "CONGESTION_SLOW_TRAFFIC"},
	"HEAVY_TRAFFIC":      {"CONGESTION", "CONGESTION_HEAVY_TRAFFIC"},
	"TRAFFIC_BUILDUP":    {"CONGESTION", "CONGESTION_TRAFFIC_BUILDING_UP"},
	"ACCIDENT":           {"INCIDENT", "INCIDENT_ACCIDENT"},
	"ROADWORKS":          {"CONSTRUCTION", "CONSTRUCTION_ROADWORKS"},
	"BLOCKAGE":           {"RESTRICTION", "RESTRICTION_BLOCKED"},
	"HAZARD":             {"HAZARD", "HAZARD_OBSTRUCTION"},
	// Add more mappings as needed
}

// translateToTraFFType translates Autobahn API event types to TraFF event types.
func translateToTraFFType(apiType string) (class string, eventType string, ok bool) {
	mapping, exists := autobahnToTraFF[apiType]
	if exists {
		return mapping.Class, mapping.Type, true
	}
	return "", "", false
}

// ConvertToTraFF converts filtered warnings into the TraFF XML format.
func ConvertToTraFF(warnings []Warning) (string, error) {
	traff := TraFF{}

	for _, warning := range warnings {
		class, eventType, ok := translateToTraFFType(warning.AbnormalTrafficType)
		if !ok {
			log.Printf("Unknown Autobahn API type: %s", warning.AbnormalTrafficType)
			continue // Skip unknown types
		}

		// Parse the 'point' field to extract latitude and longitude.
		var fromLat, fromLon float64
		n, err := fmt.Sscanf(warning.Point, "%f,%f", &fromLat, &fromLon)
		if err != nil || n != 2 {
			log.Printf("Failed to parse point: %s", warning.Point)
			continue // Skip if point parsing fails
		}

		// Construct the polyline coordinates.
		var polyline Polyline
		for _, coord := range warning.Geometry.Coordinates {
			if len(coord) != 2 {
				log.Printf("Invalid coordinate pair: %v", coord)
				continue
			}
			lon, lat := coord[0], coord[1]
			polyline.Coordinates = append(polyline.Coordinates, Coordinate{Lon: lon, Lat: lat})
		}

		// Determine the 'to' point from the last coordinate.
		var toPoint Point
		if len(polyline.Coordinates) > 0 {
			lastCoord := polyline.Coordinates[len(polyline.Coordinates)-1]
			toPoint = Point{Lat: lastCoord.Lat, Lon: lastCoord.Lon}
		}

		// Get the first description if available.
		description := "No description available"
		if len(warning.Description) > 0 && warning.Description[0] != "" {
			description = warning.Description[0]
		}

		// Create the Event.
		event := Event{
			Class:       class,
			Type:        eventType,
			Speed:       warning.AverageSpeed,
			Description: description,
		}

		// Create the Message.
		message := Message{
			Location: Location{
				From:     Point{Lat: fromLat, Lon: fromLon},
				To:       toPoint,
				Polyline: polyline,
				// JunctionRef and JunctionName can be set here if available.
			},
			Events: []Event{event},
		}

		// Append the message to the TraFF structure.
		traff.Messages = append(traff.Messages, message)
	}

	// Serialize the TraFF structure to XML.
	output, err := xml.MarshalIndent(traff, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize XML: %w", err)
	}

	return string(output), nil
}

func main() {
	// Fetch warnings from the Autobahn API.
	warnings, err := FetchWarnings()
	if err != nil {
		log.Fatalf("Error fetching warnings: %v", err)
	}

	// Convert the warnings to TraFF XML.
	xmlOutput, err := ConvertToTraFF(warnings)
	if err != nil {
		log.Fatalf("Error converting to TraFF: %v", err)
	}

	// Print the XML output.
	fmt.Println(xmlOutput)
}
