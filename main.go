package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/apognu/gocal"
)

type Event struct {
	Summary      string
	TimesPerYear int
	Rrule        map[string]string
}

type ByTimesPerYear []Event

func (e ByTimesPerYear) Less(i, j int) bool {
	a, b := e[i], e[j]
	if a.TimesPerYear != b.TimesPerYear {
		return a.TimesPerYear < b.TimesPerYear
	}
	return a.Summary < b.Summary
}

func (e ByTimesPerYear) Len() int      { return len(e) }
func (e ByTimesPerYear) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

func main() {
	var (
		fMinwidth    = flag.Int("min-width", 42, "Minimum width of the title column")
		fICSFilepath = flag.String("ics-filepath", "", "Path to the ICS File to parse")
	)
	flag.Parse()

	start, end := time.Now(), time.Now().AddDate(1, 0, 0)

	f, err := os.Open(*fICSFilepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	c := gocal.NewParser(f)
	c.Start, c.End = &start, &end
	c.Parse()

	recurringEvents := make(map[string]gocal.Event)
	freq := make(map[string]int)
	for _, e := range c.Events {
		if e.IsRecurring {
			if _, ok := recurringEvents[e.Summary]; !ok {
				recurringEvents[e.Summary] = e
			}
			freq[e.Summary] += 1
		}
	}

	// clean out emojis and such for tabwriter not to bug out
	reg, err := regexp.Compile("[[:^ascii:]]")
	if err != nil {
		log.Fatal(err)
	}

	var events []Event
	for _, ev := range recurringEvents {
		events = append(events, Event{
			Summary:      reg.ReplaceAllString(ev.Summary, ""),
			Rrule:        ev.RecurrenceRule,
			TimesPerYear: freq[ev.Summary],
		})
	}
	sort.Sort(ByTimesPerYear(events))

	minwidth := *fMinwidth
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)
	fmt.Fprintln(w, "Name\t# of times\tRRULE\t")
	for _, e := range events {
		title := e.Summary
		if len(e.Summary) > minwidth {
			title = e.Summary[0:minwidth]
		}
		fmt.Fprintf(w, "%s\t%d Times\t%s\n", title, e.TimesPerYear, e.Rrule)
	}
	w.Flush()
}
