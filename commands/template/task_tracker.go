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
		tracks map[string]*track
	}

	track struct {
		start    time.Time
		duration float64
	}
)

// NewTaskTracker create a new tracker
func NewTaskTracker() *TaskTracker {
	return &TaskTracker{
		tracks: make(map[string]*track),
	}
}

// Track the duration of the task
func (t *TaskTracker) Track(name string) {
	t.tracks[name] = &track{start: time.Now()}
}

// UnTrack measure the duration in seconds
func (t *TaskTracker) UnTrack(name string) {
	if v, ok := t.tracks[name]; ok {
		v.duration = time.Since(v.start).Seconds()
	}
}

// PrintSummary print the summary on stdout
func (t *TaskTracker) PrintSummary(output io.Writer) {
	var totalDuration float64

	for _, t := range t.tracks {
		totalDuration += t.duration
	}

	var headline, column string

	for name, t := range t.tracks {
		headline += fmt.Sprintf("%s\t", name)
		column += fmt.Sprintf("%s sec\t", strconv.FormatFloat(t.duration, 'f', 2, 64))
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
