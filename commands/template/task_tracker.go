package template

import (
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"
	"time"
)

type (
	// TaskTracker help you to track the execution time of tasks and
	// generate a summary for the cli
	TaskTracker struct {
		tracks []*track
	}

	track struct {
		name     string
		start    time.Time
		duration float64
	}
)

// NewTaskTracker create a new tracker
func NewTaskTracker() *TaskTracker {
	return &TaskTracker{
		tracks: []*track{},
	}
}

// Track the duration of the task
func (t *TaskTracker) Track(name string) {
	t.tracks = append(t.tracks, &track{name, time.Now(), 0})
}

// UnTrack measure the duration in seconds
func (t *TaskTracker) UnTrack(name string) {
	for _, v := range t.tracks {
		if v.name == name {
			v.duration = time.Since(v.start).Seconds()
			break
		}
	}
}

// PrintSummary print the summary on stdout
func (t *TaskTracker) PrintSummary(output io.Writer) {
	var totalDuration float64

	for _, t := range t.tracks {
		totalDuration += t.duration
	}

	var headline, column string

	for _, v := range t.tracks {
		headline += fmt.Sprintf("%s\t", v.name)
		column += fmt.Sprintf("%s sec\t", strconv.FormatFloat(v.duration, 'f', 2, 64))
	}

	headline += fmt.Sprintf("Total\t")
	column += fmt.Sprintf("%s sec\t", strconv.FormatFloat(totalDuration, 'f', 2, 64))

	w := new(tabwriter.Writer)
	w.Init(output, 0, 4, 2, ' ', tabwriter.StripEscape)
	fmt.Fprintln(w, headline)
	fmt.Fprintf(w, column)
	fmt.Fprintln(w)

	w.Flush()
}
