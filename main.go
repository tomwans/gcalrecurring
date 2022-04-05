package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/apognu/gocal"
)

type Event struct {
	Summary      string
	Frequency    string
	Interval     int64
	Until        string
	TimesPerYear int64
	Rrule        map[string]string
}

var freqs = map[string]int64{
	"YEARLY":  1,
	"MONTHLY": 12,
	"WEEKLY":  52,
	"DAILY":   365,
}

type ByFrequency []Event

func (e ByFrequency) Less(i, j int) bool {
	a, b := e[i], e[j]
	if a.TimesPerYear != b.TimesPerYear {
		return a.TimesPerYear < b.TimesPerYear
	}
	if a.Summary != b.Summary {
		return a.Summary < b.Summary
	}
	return a.Until < b.Until
}

func (e ByFrequency) Len() int      { return len(e) }
func (e ByFrequency) Swap(i, j int) { e[i], e[j] = e[j], e[i] }

func main() {
	var (
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

	reg, err := regexp.Compile("[^a-zA-Z0-9 /\\[\\]]+")
	if err != nil {
		log.Fatal(err)
	}

	var events []Event
	for _, ev := range recurringEvents {
		i := int64(1)
		intervalStr := ev.RecurrenceRule["INTERVAL"]
		if intervalStr != "" {
			i, err = strconv.ParseInt(intervalStr, 10, 64)
			if err != nil {
				log.Fatal(err)
			}
		}
		events = append(events, Event{
			Summary:      reg.ReplaceAllString(ev.Summary, ""),
			Rrule:        ev.RecurrenceRule,
			Frequency:    ev.RecurrenceRule["FREQ"],
			Interval:     i,
			Until:        ev.RecurrenceRule["UNTIL"],
			TimesPerYear: int64(freq[ev.Summary]),
		})
	}
	sort.Sort(ByFrequency(events))

	minwidth := 30
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)
	fmt.Fprintln(w, "Name\tFrequency\tRRULE\t")
	for _, e := range events {
		title := e.Summary
		if len(e.Summary) > minwidth {
			title = e.Summary[0:minwidth]
		}
		fmt.Fprintf(w, "%s\t%d Times\t%s\n", title, e.TimesPerYear, e.Rrule)
	}
	w.Flush()
}
