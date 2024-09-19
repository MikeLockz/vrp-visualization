package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type VRP struct {
	Problem struct {
		Description struct {
			ShiftTeams []ShiftTeam `json:"shift_teams"`
		} `json:"description"`
	} `json:"problem"`
}

type ShiftTeam struct {
	ID                  string `json:"id"`
	DepotLocationID     string `json:"depot_location_id"`
	AvailableTimeWindow struct {
		StartTimestampSec string `json:"start_timestamp_sec"`
		EndTimestampSec   string `json:"end_timestamp_sec"`
	} `json:"available_time_window"`
	RouteHistory struct {
		CurrentPosition struct {
			LocationID        string `json:"location_id"`
			KnownTimestampSec string `json:"known_timestamp_sec"`
		} `json:"current_position"`
		Stops []Stop `json:"stops"`
	} `json:"route_history"`
	UpcomingCommitments map[string]interface{} `json:"upcoming_commitments"`
	NumDHMTMembers      int                    `json:"num_dhmt_members"`
	NumAppMembers       int                    `json:"num_app_members"`
}

type Stop struct {
	Visit                        *Visit     `json:"visit,omitempty"`
	RestBreak                    *RestBreak `json:"rest_break,omitempty"`
	Pinned                       bool       `json:"pinned"`
	ActualStartTimestampSec      string     `json:"actual_start_timestamp_sec,omitempty"`
	ActualCompletionTimestampSec string     `json:"actual_completion_timestamp_sec,omitempty"`
}

type Visit struct {
	VisitID             string `json:"visit_id"`
	ArrivalTimestampSec string `json:"arrival_timestamp_sec,omitempty"`
}

type RestBreak struct {
	RestBreakID        string `json:"rest_break_id"`
	StartTimestampSec  string `json:"start_timestamp_sec"`
}

func convertTimestamp(timestamp string) time.Time {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}
	}
	t := time.Unix(i, 0)
	
	mountainTime, err := time.LoadLocation("America/Denver")
	if err != nil {
		return time.Time{}
	}
	
	return t.In(mountainTime)
}

func formatTime(t time.Time) string {
	return t.Format("15:04:05")
}

func generateTimeline(stops []Stop) string {
	if len(stops) == 0 {
		return "No stops in the timeline."
	}

	// Sort stops by start time
	sort.Slice(stops, func(i, j int) bool {
		timeI := getStopTime(stops[i])
		timeJ := getStopTime(stops[j])
		return timeI.Before(timeJ)
	})

	startTime := getStopTime(stops[0])
	endTime := getStopTime(stops[len(stops)-1])
	duration := endTime.Sub(startTime)

	timelineWidth := 80
	var timeline strings.Builder

	for _, stop := range stops {
		stopTime := getStopTime(stop)
		relativePosition := float64(stopTime.Sub(startTime)) / float64(duration)
		position := int(relativePosition * float64(timelineWidth))

		timeline.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", position), getStopMarker(stop)))
		timeline.WriteString(fmt.Sprintf("%s|\n", strings.Repeat(" ", position)))
	}

	// Add time labels
	timeline.WriteString(fmt.Sprintf("%-*s%s\n", timelineWidth, formatTime(startTime), formatTime(endTime)))

	return timeline.String()
}

func getStopTime(stop Stop) time.Time {
	if stop.Visit != nil && stop.Visit.ArrivalTimestampSec != "" {
		return convertTimestamp(stop.Visit.ArrivalTimestampSec)
	} else if stop.RestBreak != nil {
		return convertTimestamp(stop.RestBreak.StartTimestampSec)
	}
	return time.Time{}
}

func getStopMarker(stop Stop) string {
	if stop.Visit != nil {
		return fmt.Sprintf("V%s", stop.Visit.VisitID)
	} else if stop.RestBreak != nil {
		return fmt.Sprintf("B%s", stop.RestBreak.RestBreakID)
	}
	return "?"
}

func main() {
	// ... (file opening and JSON decoding remain the same)

	for _, team := range vrp.Problem.Description.ShiftTeams {
		fmt.Printf("Shift Team ID: %s\n", team.ID)
		fmt.Printf("Depot Location ID: %s\n", team.DepotLocationID)
		fmt.Printf("Available Time Window: %s - %s\n",
			formatTime(convertTimestamp(team.AvailableTimeWindow.StartTimestampSec)),
			formatTime(convertTimestamp(team.AvailableTimeWindow.EndTimestampSec)))
		
		fmt.Printf("Current Position: Location ID %s, Known Time %s\n",
			team.RouteHistory.CurrentPosition.LocationID,
			formatTime(convertTimestamp(team.RouteHistory.CurrentPosition.KnownTimestampSec)))
		
		fmt.Println("\nRoute History Timeline:")
		fmt.Println(generateTimeline(team.RouteHistory.Stops))
		
		fmt.Println("\nDetailed Route History:")
		for i, stop := range team.RouteHistory.Stops {
			fmt.Printf("  Stop %d:\n", i+1)
			if stop.Visit != nil {
				fmt.Printf("    Visit ID: %s\n", stop.Visit.VisitID)
				if stop.Visit.ArrivalTimestampSec != "" {
					fmt.Printf("    Arrival Time: %s\n", formatTime(convertTimestamp(stop.Visit.ArrivalTimestampSec)))
				}
			} else if stop.RestBreak != nil {
				fmt.Printf("    Rest Break ID: %s\n", stop.RestBreak.RestBreakID)
				fmt.Printf("    Start Time: %s\n", formatTime(convertTimestamp(stop.RestBreak.StartTimestampSec)))
			}
			fmt.Printf("    Pinned: %v\n", stop.Pinned)
			if stop.ActualStartTimestampSec != "" {
				fmt.Printf("    Actual Start: %s\n", formatTime(convertTimestamp(stop.ActualStartTimestampSec)))
			}
			if stop.ActualCompletionTimestampSec != "" {
				fmt.Printf("    Actual Completion: %s\n", formatTime(convertTimestamp(stop.ActualCompletionTimestampSec)))
			}
		}
		
		fmt.Printf("Number of DHMT Members: %d\n", team.NumDHMTMembers)
		fmt.Printf("Number of App Members: %d\n", team.NumAppMembers)
		fmt.Println()
	}
}