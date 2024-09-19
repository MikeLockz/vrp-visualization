package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

func convertTimestamp(timestamp string) string {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "Invalid Timestamp"
	}
	t := time.Unix(i, 0)
	
	// Load the Mountain Time location
	mountainTime, err := time.LoadLocation("America/Denver")
	if err != nil {
		return "Error loading timezone"
	}
	
	// Convert to Mountain Time
	t = t.In(mountainTime)
	
	// Format as HH:mm:ss
	return t.Format("15:04:05")
}

func main() {
	// Open the JSON file
	file, err := os.Open("vrp_problem.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Create a decoder
	decoder := json.NewDecoder(file)

	// Create a VRP struct to hold the data
	var vrp VRP

	// Decode the JSON data
	err = decoder.Decode(&vrp)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// Print details for all shift teams
	for _, team := range vrp.Problem.Description.ShiftTeams {
		fmt.Printf("Shift Team ID: %s\n", team.ID)
		fmt.Printf("Depot Location ID: %s\n", team.DepotLocationID)
		fmt.Printf("Available Time Window: %s - %s\n",
			convertTimestamp(team.AvailableTimeWindow.StartTimestampSec),
			convertTimestamp(team.AvailableTimeWindow.EndTimestampSec))
		
		fmt.Printf("Current Position: Location ID %s, Known Time %s\n",
			team.RouteHistory.CurrentPosition.LocationID,
			convertTimestamp(team.RouteHistory.CurrentPosition.KnownTimestampSec))
		
		fmt.Println("Route History:")
		for i, stop := range team.RouteHistory.Stops {
			fmt.Printf("  Stop %d:\n", i+1)
			if stop.Visit != nil {
				fmt.Printf("    Visit ID: %s\n", stop.Visit.VisitID)
				if stop.Visit.ArrivalTimestampSec != "" {
					fmt.Printf("    Arrival Time: %s\n", convertTimestamp(stop.Visit.ArrivalTimestampSec))
				}
			} else if stop.RestBreak != nil {
				fmt.Printf("    Rest Break ID: %s\n", stop.RestBreak.RestBreakID)
				fmt.Printf("    Start Time: %s\n", convertTimestamp(stop.RestBreak.StartTimestampSec))
			}
			fmt.Printf("    Pinned: %v\n", stop.Pinned)
			if stop.ActualStartTimestampSec != "" {
				fmt.Printf("    Actual Start: %s\n", convertTimestamp(stop.ActualStartTimestampSec))
			}
			if stop.ActualCompletionTimestampSec != "" {
				fmt.Printf("    Actual Completion: %s\n", convertTimestamp(stop.ActualCompletionTimestampSec))
			}
		}
		
		fmt.Printf("Number of DHMT Members: %d\n", team.NumDHMTMembers)
		fmt.Printf("Number of App Members: %d\n", team.NumAppMembers)
		fmt.Println()
	}
}