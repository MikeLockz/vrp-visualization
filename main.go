package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type Problem struct {
	Problem struct {
		Description struct {
			ShiftTeams []struct {
				ID                 string `json:"id"`
				AvailableTimeWindow struct {
					StartTimestampSec string `json:"start_timestamp_sec"`
					EndTimestampSec   string `json:"end_timestamp_sec"`
				} `json:"available_time_window"`
				RouteHistory struct {
					Stops []struct {
						Visit struct {
							VisitID string `json:"visit_id"`
						} `json:"visit"`
						RestBreak struct {
							RestBreakID string `json:"rest_break_id"`
						} `json:"rest_break"`
						ActualStartTimestampSec      string `json:"actual_start_timestamp_sec"`
						ActualCompletionTimestampSec string `json:"actual_completion_timestamp_sec"`
					} `json:"stops"`
				} `json:"route_history"`
			} `json:"shift_teams"`
			Visits []struct {
				ID                string `json:"id"`
				ArrivalTimeWindow struct {
					StartTimestampSec string `json:"start_timestamp_sec"`
					EndTimestampSec   string `json:"end_timestamp_sec"`
				} `json:"arrival_time_window"`
			} `json:"visits"`
			RestBreaks []struct {
				ID           string `json:"id"`
				ShiftTeamID  string `json:"shift_team_id"`
				DurationSec  string `json:"duration_sec"`
				Unrequested  bool   `json:"unrequested"`
				LocationID   string `json:"location_id"`
				StartTimestampSec string `json:"start_timestamp_sec"`
			} `json:"rest_breaks"`
		} `json:"description"`
	} `json:"problem"`
}

func main() {
	// Read the JSON file
	data, err := ioutil.ReadFile("vrp.json")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Parse the JSON data
	var problem Problem
	err = json.Unmarshal(data, &problem)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Load Mountain Time location
	mountainTime, err := time.LoadLocation("America/Denver")
	if err != nil {
		log.Fatalf("Error loading Mountain Time zone: %v", err)
	}

	// Create a map of visit IDs to their arrival time windows
	visitWindows := make(map[string]struct {
		Start string
		End   string
	})
	for _, visit := range problem.Problem.Description.Visits {
		visitWindows[visit.ID] = struct {
			Start string
			End   string
		}{
			Start: formatTimestamp(visit.ArrivalTimeWindow.StartTimestampSec, mountainTime),
			End:   formatTimestamp(visit.ArrivalTimeWindow.EndTimestampSec, mountainTime),
		}
	}

	// Create a map of break IDs to their details
	breakDetails := make(map[string]struct {
		Duration string
		Start    string
	})
	for _, restBreak := range problem.Problem.Description.RestBreaks {
		duration, _ := strconv.Atoi(restBreak.DurationSec)
		breakDetails[restBreak.ID] = struct {
			Duration string
			Start    string
		}{
			Duration: formatDuration(duration),
			Start:    formatTimestamp(restBreak.StartTimestampSec, mountainTime),
		}
	}

	// Process each shift team
	for _, team := range problem.Problem.Description.ShiftTeams {
		fmt.Printf("Shift Team ID: %s\n", team.ID)
		
		// Print shift team start and end times
		shiftStart := formatTimestamp(team.AvailableTimeWindow.StartTimestampSec, mountainTime)
		shiftEnd := formatTimestamp(team.AvailableTimeWindow.EndTimestampSec, mountainTime)
		fmt.Printf("Shift Start Time: %s\n", shiftStart)
		fmt.Printf("Shift End Time: %s\n", shiftEnd)
		
		for _, stop := range team.RouteHistory.Stops {
			if stop.Visit.VisitID != "" {
				fmt.Printf("  Visit ID: %s\n", stop.Visit.VisitID)
				if window, exists := visitWindows[stop.Visit.VisitID]; exists {
					fmt.Printf("    Arrival Window Start: %s\n", window.Start)
					fmt.Printf("    Arrival Window End: %s\n", window.End)
				}
			} else if stop.RestBreak.RestBreakID != "" {
				fmt.Printf("  Break ID: %s\n", stop.RestBreak.RestBreakID)
				if details, exists := breakDetails[stop.RestBreak.RestBreakID]; exists {
					fmt.Printf("    Scheduled Start: %s\n", details.Start)
					fmt.Printf("    Duration: %s\n", details.Duration)
				}
			} else {
				fmt.Printf("  Unknown Stop Type\n")
			}

			startTime := formatTimestamp(stop.ActualStartTimestampSec, mountainTime)
			endTime := formatTimestamp(stop.ActualCompletionTimestampSec, mountainTime)
			fmt.Printf("    Actual Start Time: %s\n", startTime)
			fmt.Printf("    Actual End Time: %s\n", endTime)
		}
		fmt.Println()
	}
}

func formatTimestamp(timestamp string, loc *time.Location) string {
	if timestamp == "" {
		return "Not available"
	}
	
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "Invalid timestamp"
	}
	
	t := time.Unix(i, 0).In(loc)
	return t.Format("2006-01-02 03:04:05 PM MST")
}

func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}