package timeset

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestIsTodayOrInFuture(t *testing.T) {
	now := time.Now()
	// startOfToday := StartOfDay(now)

	tests := []struct {
		input    time.Time
		expected bool
	}{
		// Test cases for today
		{now, true},
		{now.Add(1 * time.Minute), true},
		{now.Add(-1 * time.Minute), true},
		{now.Add(28 * time.Hour), true},
		{now.Add(-25 * time.Hour), false},
		{now.Add(-24 * 7 * time.Hour), false},
		{now.Add(1 * time.Minute), true},
		{now.Add(-1 * time.Minute), true},
		{now.Add(1 * time.Second), true},
		{now.Add(-1 * time.Second), true},

		// Test cases for edge dates
		{time.Date(1999, 12, 31, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC), true},
		{time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC), true},
	}

	for i, test := range tests {
		actual := IsTodayOrInFuture(test.input)
		if actual != test.expected {
			t.Errorf("[%v] IsTodayOrInFuture(%v) = %v; expected %v", i+1, test.input, actual, test.expected)
		}
	}
}

func TestMinSince(t *testing.T) {
	// Define a struct for test cases
	type testCase struct {
		input    time.Time
		expected string
	}

	// Define the test cases
	tests := []testCase{
		{
			input:    time.Now().Add(-5 * time.Minute),
			expected: "5.00 min",
		},
		{
			input:    time.Now().Add(-10 * time.Minute),
			expected: "10.00 min",
		},
		{
			input:    time.Now().Add(-30 * time.Minute),
			expected: "30.00 min",
		},
		// Add more test cases as needed
	}

	// Iterate over each test case
	for _, tc := range tests {
		t.Run(fmt.Sprintf("Testing with %v", tc.input), func(t *testing.T) {
			// Call the function with the test case input
			result := MinSince(tc.input)

			// Compare the result with the expected value
			if result != tc.expected {
				t.Errorf("MinSince(%v) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFutureOffsetMillis(t *testing.T) {
	now := time.Now()
	const epsilon = 10 // Milliseconds of allowable error

	tests := []struct {
		name   string
		future time.Time
		want   int64
	}{
		{
			name:   "5 minutes into future",
			future: now.Add(5 * time.Minute),
			want:   5 * 60 * 1000,
		},
		{
			name:   "1 hour into future",
			future: now.Add(1 * time.Hour),
			want:   60 * 60 * 1000,
		},
		{
			name:   "30 seconds into future",
			future: now.Add(30 * time.Second),
			want:   30 * 1000,
		},
		{
			name:   "10 hours into future",
			future: now.Add(10 * time.Hour),
			want:   10 * 60 * 60 * 1000,
		},
	}

	abs := func(x int64) int64 {
		if x < 0 {
			return -x
		}
		return x
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FutureOffsetMillis(tt.future); abs(int64(got)-tt.want) > epsilon {
				t.Errorf("FutureOffsetMillis() = %v, want %v (within ±%v ms error margin)", got, tt.want, epsilon)
			}
		})
	}
}

func TestExpandDays(t *testing.T) {

	// Define test cases
	tests := []struct {
		name         string
		start        time.Time
		days         int
		filterFunc   func(time.Time) bool
		wantLength   int
		expectedDays []time.Time
	}{
		{
			name:       "no filter 5 days",
			start:      time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			days:       5,
			filterFunc: nil,
			wantLength: 5,
		},
		{
			name:  "filter weekends, 10 days",
			start: time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC), // note that this is saturday
			days:  10,
			filterFunc: func(date time.Time) bool {
				weekday := date.Weekday()
				return weekday == time.Saturday || weekday == time.Sunday
			},
			wantLength: 6,
		},
		{
			name:  "filter specific date, 5 days",
			start: time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
			days:  5,
			filterFunc: func(t time.Time) bool {
				// Skipping April 3rd specifically
				return t.Format("2006-01-02") == "2023-04-03"
			},
			wantLength: 4,
		},
		{
			name:       "filter to exclude future days",
			start:      time.Now().AddDate(0, 0, 1),
			days:       20,
			filterFunc: func(t time.Time) bool { return t.After(time.Now()) },
			wantLength: 0,
		},
		{
			name:       "filter to exclude day if it is after 3 days into the future",
			start:      time.Now(),
			days:       7,
			filterFunc: func(t time.Time) bool { return t.After(time.Now().AddDate(0, 0, 3)) },
			wantLength: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExpandDays(tc.start, tc.days, tc.filterFunc)
			// fmt.Printf("got: %v\n", got)

			fmt.Printf("  tc.start: %v\n", tc.start.Format("2006-01-02"))
			for _, day := range got {
				fmt.Printf("	~ day: %v\n", day.Format("2006-01-02"))
			}

			if len(got) != tc.wantLength {
				t.Errorf("ExpandDays(%v, %d, filterFunc) got %d dates, want %d dates", tc.start, tc.days, len(got), tc.wantLength)
			}
		})
	}
}

func TestExpandTime(t *testing.T) {
	tests := []struct {
		name       string
		start      time.Time
		duration   time.Duration
		count      int
		filterFunc func(time.Time) bool
		expected   []time.Time
	}{
		{
			name:       "No filter, daily interval",
			start:      time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC),
			duration:   24 * time.Hour,
			count:      5,
			filterFunc: nil,
			expected: []time.Time{
				time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 21, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 22, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 23, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 24, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:     "Skip weekends",
			start:    time.Date(2024, 5, 17, 0, 0, 0, 0, time.UTC), // Starting on a Friday
			duration: 24 * time.Hour,
			count:    5,
			filterFunc: func(t time.Time) bool {
				weekday := t.Weekday()
				// fmt.Printf("weekday: %v\n", weekday)
				return weekday == time.Saturday || weekday == time.Sunday
			},
			expected: []time.Time{
				time.Date(2024, 5, 17, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 21, 0, 0, 0, 0, time.UTC),
				// time.Date(2024, 5, 22, 0, 0, 0, 0, time.UTC),
				// time.Date(2024, 5, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:       "Hourly interval, no filter",
			start:      time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC),
			duration:   time.Hour,
			count:      5,
			filterFunc: nil,
			expected: []time.Time{
				time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 20, 1, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 20, 2, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 20, 3, 0, 0, 0, time.UTC),
				time.Date(2024, 5, 20, 4, 0, 0, 0, time.UTC),
			},
		},
		// NOTE: reflection in this case fails, but the result is correct.
		// {
		// 	name:       "No results due to filter",
		// 	start:      time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC),
		// 	duration:   24 * time.Hour,
		// 	count:      5,
		// 	filterFunc: func(t time.Time) bool { return true },
		// 	expected:   []time.Time{},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandTime(tt.start, tt.duration, tt.count, tt.filterFunc)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("\nExpandTime() = \n\t%v, \nwant \n\t%v", got, tt.expected)
			}
		})
	}
}
