package utils

// Provides duration utilities.

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// DurationMillis returns the milliseconds value for the provided duration.
func DurationMillis(dur time.Duration) int64 {
	return dur.Nanoseconds() / 1e6
}

// ElapsedMillis returns the elapsed time between start and end in milliseconds.
// Use variadic to NOT require end, in which case "now" is assumed.
func ElapsedMillis(start time.Time, ends ...time.Time) int64 {
	var end time.Time
	switch len(ends) {
	case 0:
		end = time.Now()
	case 1:
		end = ends[0]
	default:
		log.WithFields(log.Fields{"start": start, "ends": ends}).Panic("Invalid multiple ends for ElapsedMillis")
	}
	return DurationMillis(end.Sub(start))
}

// SecsToDuration converts seconds from int to Duration.
func SecsToDuration(secs int) time.Duration {
	return time.Duration(secs) * time.Second
}

const minimumWeekOfYear = 1

// FirstDayOfISOWeek returns the time instance of the first day of the given ISO week in given timezone.
func FirstDayOfISOWeek(year int, week int, timezone *time.Location) time.Time {
	if week < minimumWeekOfYear {
		week = minimumWeekOfYear
	}

	maximumWeekOfYear := getMaximumWeekOfYear(year, timezone)
	if week > maximumWeekOfYear {
		week = maximumWeekOfYear
	}

	date := time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()
	for date.Weekday() != time.Monday { // iterate back to Monday
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 1)
		_, isoWeek = date.ISOWeek()
	}
	return date
}

func getMaximumWeekOfYear(year int, timezone *time.Location) int {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, timezone)

	isoYear, isoWeek := date.ISOWeek()

	var prevWeek int
	for isoYear <= year {
		prevWeek = isoWeek
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}

	return prevWeek
}

// GetWeekStartTime returns week start (Monday) timestamp of the given ts in given timezone.
func GetWeekStartTime(ts int64, timezone *time.Location) time.Time {
	year, month, day := time.Unix(ts, 0).In(timezone).Date()
	zoneDate := time.Date(year, month, day, 0, 0, 0, 0, timezone)
	zoneYear, zoneWeek := zoneDate.ISOWeek()
	return FirstDayOfISOWeek(zoneYear, zoneWeek, timezone)
}

// GetDateStartTime returns date start timestamp (at 00:00:00) of the given ts in given timezone.
func GetDateStartTime(ts int64, timezone *time.Location) time.Time {
	zoneYear, zoneMonth, zoneDay := time.Unix(ts, 0).In(timezone).Date()
	return time.Date(zoneYear, zoneMonth, zoneDay, 0, 0, 0, 0, timezone)
}
