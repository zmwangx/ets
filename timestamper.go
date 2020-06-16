package main

import (
	"bytes"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/lestrrat-go/strftime"
)

type TimestampMode int

const (
	AbsoluteTimeMode TimestampMode = iota
	ElapsedTimeMode
	IncrementalTimeMode
)

type Timestamper struct {
	Mode           TimestampMode
	TZ             *time.Location
	Formatter      *strftime.Strftime
	StartTimestamp time.Time
	LastTimestamp  time.Time
}

func NewTimestamper(format string, mode TimestampMode, timezone *time.Location) (*Timestamper, error) {
	formatter, err := strftime.New(format,
		strftime.WithMilliseconds('L'),
		strftime.WithUnixSeconds('s'),
		strftime.WithSpecification('f', microseconds))
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Timestamper{
		Mode:           mode,
		TZ:             timezone,
		Formatter:      formatter,
		StartTimestamp: now,
		LastTimestamp:  now,
	}, nil
}

func (t *Timestamper) CurrentTimestampString() string {
	now := time.Now()
	var s string
	switch t.Mode {
	case AbsoluteTimeMode:
		s = t.Formatter.FormatString(time.Now().In(t.TZ))
	case ElapsedTimeMode:
		s = formatDuration(t.Formatter, now.Sub(t.StartTimestamp))
	case IncrementalTimeMode:
		s = formatDuration(t.Formatter, now.Sub(t.LastTimestamp))
	default:
		log.Panic("unknown mode ", t.Mode)
	}
	t.LastTimestamp = now
	return s
}

func formatDuration(formatter *strftime.Strftime, duration time.Duration) string {
	return formatter.FormatString(time.Unix(0, duration.Nanoseconds()).UTC())
}

var microseconds strftime.Appender

func init() {
	microseconds = strftime.AppendFunc(func(b []byte, t time.Time) []byte {
		microsecond := int(t.Nanosecond()) / int(time.Microsecond)
		if microsecond == 0 {
			return append(b, "000000"...)
		} else {
			length := int(math.Log10(float64(microsecond))) + 1
			b = append(b, bytes.Repeat([]byte("0"), 6-length)...)
			return append(b, strconv.Itoa(microsecond)...)
		}
	})
}
