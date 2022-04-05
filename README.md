# gcalrecurring

The G stands for "Good."

This project examines an ICS calendar file and dumps out a table of recurring events, sorted from least frequent to most frequent, with frequency information included.

## How To Use

First, download the ICS file of the calendar you want to examine.

Then:

```bash
$ go run main.go -ics-filepath=/path/to/file.ics
```
