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
				ID                string `json:"id"`
				ShiftTeamID       string `json:"shift_team_id"`
				DurationSec       string `json:"duration_sec"`
				Unrequested       bool   `json:"unrequested"`
				LocationID        string `json:"location_id"`
				StartTimestampSec string `json:"start_timestamp_sec"`
			} `json:"rest_breaks"`
		} `json:"description"`
	} `json:"problem"`
}

const (
	svgWidth  = 1000
	svgHeight = 400
	margin    = 50
)

type Stop struct {
	ID            string
	Type          string // "visit" or "break"
	WindowStart   int64
	WindowEnd     int64
	ActualStart   int64
	ActualEnd     int64
}

func main() {
	// Read the JSON file
	data, err := ioutil.ReadFile("vrp_snapshot.json")
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
		
		shiftStart, _ := strconv.ParseInt(team.AvailableTimeWindow.StartTimestampSec, 10, 64)
		shiftEnd, _ := strconv.ParseInt(team.AvailableTimeWindow.EndTimestampSec, 10, 64)
		
		fmt.Printf("Shift Start Time: %s\n", formatTimestamp(team.AvailableTimeWindow.StartTimestampSec, mountainTime))
		fmt.Printf("Shift End Time: %s\n", formatTimestamp(team.AvailableTimeWindow.EndTimestampSec, mountainTime))
		
		stops := []Stop{}

		for _, stop := range team.RouteHistory.Stops {
			var stopInfo Stop

			if stop.Visit.VisitID != "" {
				fmt.Printf("  Visit ID: %s\n", stop.Visit.VisitID)
				stopInfo.ID = stop.Visit.VisitID
				stopInfo.Type = "visit"
				if window, exists := visitWindows[stop.Visit.VisitID]; exists {
					fmt.Printf("    Arrival Window Start: %s\n", window.Start)
					fmt.Printf("    Arrival Window End: %s\n", window.End)
					visitIndex := findVisitIndex(problem.Problem.Description.Visits, stop.Visit.VisitID)
					if visitIndex != -1 {
						stopInfo.WindowStart, _ = strconv.ParseInt(problem.Problem.Description.Visits[visitIndex].ArrivalTimeWindow.StartTimestampSec, 10, 64)
						stopInfo.WindowEnd, _ = strconv.ParseInt(problem.Problem.Description.Visits[visitIndex].ArrivalTimeWindow.EndTimestampSec, 10, 64)
					}
				}
				stopInfo.ActualStart, _ = strconv.ParseInt(stop.ActualStartTimestampSec, 10, 64)
				stopInfo.ActualEnd, _ = strconv.ParseInt(stop.ActualCompletionTimestampSec, 10, 64)
			} else if stop.RestBreak.RestBreakID != "" {
				fmt.Printf("  Break ID: %s\n", stop.RestBreak.RestBreakID)
				stopInfo.ID = stop.RestBreak.RestBreakID
				stopInfo.Type = "break"
				if details, exists := breakDetails[stop.RestBreak.RestBreakID]; exists {
					fmt.Printf("    Scheduled Start: %s\n", details.Start)
					fmt.Printf("    Duration: %s\n", details.Duration)
					breakIndex := findBreakIndex(problem.Problem.Description.RestBreaks, stop.RestBreak.RestBreakID)
					if breakIndex != -1 {
						stopInfo.WindowStart, _ = strconv.ParseInt(problem.Problem.Description.RestBreaks[breakIndex].StartTimestampSec, 10, 64)
						duration, _ := strconv.ParseInt(problem.Problem.Description.RestBreaks[breakIndex].DurationSec, 10, 64)
						stopInfo.WindowEnd = stopInfo.WindowStart + duration
						
						// For breaks, set ActualStart and ActualEnd to be the same as WindowStart and WindowEnd
						stopInfo.ActualStart = stopInfo.WindowStart
						stopInfo.ActualEnd = stopInfo.WindowEnd
					}
				}
			} else {
				fmt.Printf("  Unknown Stop Type\n")
				continue
			}

			startTime := formatTimestamp(strconv.FormatInt(stopInfo.ActualStart, 10), mountainTime)
			endTime := formatTimestamp(strconv.FormatInt(stopInfo.ActualEnd, 10), mountainTime)
			fmt.Printf("    Actual Start Time: %s\n", startTime)
			fmt.Printf("    Actual End Time: %s\n", endTime)

			stops = append(stops, stopInfo)
		}
		fmt.Println()

		// Generate SVG
		svgContent := generateSVG(stops, shiftStart, shiftEnd, mountainTime)
		
		// Write SVG to file
		err := ioutil.WriteFile(fmt.Sprintf("timeline_%s.svg", team.ID), []byte(svgContent), 0644)
		if err != nil {
			fmt.Printf("Error writing SVG file: %v\n", err)
		} else {
			fmt.Printf("SVG timeline generated: timeline_%s.svg\n", team.ID)
		}
	}
}

func generateSVG(stops []Stop, shiftStart, shiftEnd int64, loc *time.Location) string {
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, svgWidth, svgHeight+100)
	
	// Define styles
	svg += `
		<style>
			.title { font: bold 24px 'Arial', sans-serif; fill: #333; }
			.legend-text { font: 14px 'Arial', sans-serif; fill: #666; }
			.axis-text { font: 12px 'Arial', sans-serif; fill: #999; }
			.label-text { font: 12px 'Arial', sans-serif; fill: #333; }
			.time-text { font: 10px 'Arial', sans-serif; fill: #666; }
			.visit-box { fill: #4CAF50; opacity: 0.7; }
			.break-box { fill: #FFC107; opacity: 0.7; }
			.window-box { fill: none; stroke: #2196F3; stroke-width: 2; opacity: 0.5; }
		</style>
	`

	// Add title
	svg += fmt.Sprintf(`<text x="%d" y="40" class="title" text-anchor="middle">Shift Team Timeline</text>`, svgWidth/2)

	// Add legend
	legendY := 70
	svg += fmt.Sprintf(`<rect x="10" y="%d" width="20" height="20" class="visit-box" rx="3" />`, legendY)
	svg += fmt.Sprintf(`<text x="40" y="%d" class="legend-text" alignment-baseline="middle">Visit</text>`, legendY+10)
	svg += fmt.Sprintf(`<rect x="100" y="%d" width="20" height="20" class="break-box" rx="3" />`, legendY)
	svg += fmt.Sprintf(`<text x="130" y="%d" class="legend-text" alignment-baseline="middle">Break</text>`, legendY+10)
	svg += fmt.Sprintf(`<rect x="190" y="%d" width="20" height="20" class="window-box" rx="3" />`, legendY)
	svg += fmt.Sprintf(`<text x="220" y="%d" class="legend-text" alignment-baseline="middle">Arrival Window</text>`, legendY+10)

	// Add background
	svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#f8f9fa" rx="5" />`, margin-10, margin+90, svgWidth-2*margin+20, svgHeight-2*margin+20)

	timeScale := float64(svgWidth-2*margin) / float64(shiftEnd-shiftStart)

	// Draw time axis
	shiftStartTime := time.Unix(shiftStart, 0).In(loc)
	for i := 0; i <= 24; i++ {
		timePoint := shiftStartTime.Add(time.Duration(i) * time.Hour)
		if timePoint.Unix() > shiftEnd {
			break
		}
		x := margin + int(float64(timePoint.Unix()-shiftStart)*timeScale)
		svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#e9ecef" stroke-width="1" />`, x, margin+100, x, svgHeight+80)
		svg += fmt.Sprintf(`<text x="%d" y="%d" class="axis-text" text-anchor="middle">%s</text>`, x, svgHeight+95, timePoint.Format("15:04"))
	}

	yOffset := margin + 120
	for i, stop := range stops {
		windowStartX := margin + int(float64(stop.WindowStart-shiftStart)*timeScale)
		windowEndX := margin + int(float64(stop.WindowEnd-shiftStart)*timeScale)
		actualStartX := margin + int(float64(stop.ActualStart-shiftStart)*timeScale)
		actualEndX := margin + int(float64(stop.ActualEnd-shiftStart)*timeScale)

		// Alternating row background
		if i%2 == 0 {
			svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="40" fill="#f1f3f5" rx="3" />`, margin-10, yOffset-10, svgWidth-2*margin+20)
		}

		if stop.Type == "break" {
			svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="20" class="break-box" rx="3" />`, actualStartX, yOffset, actualEndX-actualStartX)
		} else {
			svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="20" class="window-box" rx="3" />`, windowStartX, yOffset, windowEndX-windowStartX)
			svg += fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="20" class="visit-box" rx="3" />`, actualStartX, yOffset, actualEndX-actualStartX)
		}

		// Add label
		svg += fmt.Sprintf(`<text x="%d" y="%d" class="label-text">%s (%s)</text>`, actualStartX+5, yOffset-5, stop.ID, stop.Type)

		// Add time labels
		if stop.Type == "break" {
			actualStartTime := time.Unix(stop.ActualStart, 0).In(loc).Format("15:04")
			actualEndTime := time.Unix(stop.ActualEnd, 0).In(loc).Format("15:04")
			svg += fmt.Sprintf(`<text x="%d" y="%d" class="time-text" text-anchor="end">%s</text>`, actualStartX-5, yOffset+15, actualStartTime)
			svg += fmt.Sprintf(`<text x="%d" y="%d" class="time-text">%s</text>`, actualEndX+5, yOffset+15, actualEndTime)
		} else {
			windowStartTime := time.Unix(stop.WindowStart, 0).In(loc).Format("15:04")
			windowEndTime := time.Unix(stop.WindowEnd, 0).In(loc).Format("15:04")
			actualStartTime := time.Unix(stop.ActualStart, 0).In(loc).Format("15:04")
			actualEndTime := time.Unix(stop.ActualEnd, 0).In(loc).Format("15:04")
			svg += fmt.Sprintf(`<text x="%d" y="%d" class="time-text" text-anchor="end">%s</text>`, windowStartX-5, yOffset+30, windowStartTime)
			svg += fmt.Sprintf(`<text x="%d" y="%d" class="time-text">%s</text>`, windowEndX+5, yOffset+30, windowEndTime)
			svg += fmt.Sprintf(`<text x="%d" y="%d" class="time-text" text-anchor="end">%s</text>`, actualStartX-5, yOffset+15, actualStartTime)
			svg += fmt.Sprintf(`<text x="%d" y="%d" class="time-text">%s</text>`, actualEndX+5, yOffset+15, actualEndTime)
		}

		yOffset += 50
	}

	svg += "</svg>"
	return svg
}

func getColor(stopType string) string {
	if stopType == "visit" {
		return "green"
	}
	return "orange"
}

func findVisitIndex(visits []struct {
	ID                string `json:"id"`
	ArrivalTimeWindow struct {
		StartTimestampSec string `json:"start_timestamp_sec"`
		EndTimestampSec   string `json:"end_timestamp_sec"`
	} `json:"arrival_time_window"`
}, id string) int {
	for i, v := range visits {
		if v.ID == id {
			return i
		}
	}
	return -1
}

func findBreakIndex(breaks []struct {
	ID                string `json:"id"`
	ShiftTeamID       string `json:"shift_team_id"`
	DurationSec       string `json:"duration_sec"`
	Unrequested       bool   `json:"unrequested"`
	LocationID        string `json:"location_id"`
	StartTimestampSec string `json:"start_timestamp_sec"`
}, id string) int {
	for i, b := range breaks {
		if b.ID == id {
			return i
		}
	}
	return -1
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