package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIsMonSunSchedule(t *testing.T) {
	timeBlock := TimeBlockModel{Heating: types.Bool{}, Temperature: types.Float64{}, Start: types.String{}, End: types.String{}}
	cases := []struct {
		schedule HeatingScheduleResourceModel
		expected bool
	}{
		// valid mon-sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
			},
			expected: true,
		},
		// valid mon-fri, sat, sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonFri: []TimeBlockModel{timeBlock},
				Sat:    []TimeBlockModel{timeBlock},
				Sun:    []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// valid mon, tue, wed, thu, fri, sat, sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				Mon: []TimeBlockModel{timeBlock},
				Tue: []TimeBlockModel{timeBlock},
				Wed: []TimeBlockModel{timeBlock},
				Thu: []TimeBlockModel{timeBlock},
				Fri: []TimeBlockModel{timeBlock},
				Sat: []TimeBlockModel{timeBlock},
				Sun: []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// invalid empty schedule
		{
			schedule: HeatingScheduleResourceModel{},
			expected: false,
		},
		// invalid mixed schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
				MonFri: []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// invalid full schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
				MonFri: []TimeBlockModel{timeBlock},
				Mon:    []TimeBlockModel{timeBlock},
				Tue:    []TimeBlockModel{timeBlock},
				Wed:    []TimeBlockModel{timeBlock},
				Thu:    []TimeBlockModel{timeBlock},
				Fri:    []TimeBlockModel{timeBlock},
				Sat:    []TimeBlockModel{timeBlock},
				Sun:    []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		actual := isMonSunSchedule(c.schedule)
		if actual != c.expected {
			t.Fatalf("Expected: %t, got: %t", c.expected, actual)
		}
	}
}

func TestIsMonFriSatSunSchedule(t *testing.T) {
	timeBlock := TimeBlockModel{Heating: types.Bool{}, Temperature: types.Float64{}, Start: types.String{}, End: types.String{}}
	cases := []struct {
		schedule HeatingScheduleResourceModel
		expected bool
	}{
		// valid mon-sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// valid mon-fri, sat, sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonFri: []TimeBlockModel{timeBlock},
				Sat:    []TimeBlockModel{timeBlock},
				Sun:    []TimeBlockModel{timeBlock},
			},
			expected: true,
		},
		// valid mon, tue, wed, thu, fri, sat, sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				Mon: []TimeBlockModel{timeBlock},
				Tue: []TimeBlockModel{timeBlock},
				Wed: []TimeBlockModel{timeBlock},
				Thu: []TimeBlockModel{timeBlock},
				Fri: []TimeBlockModel{timeBlock},
				Sat: []TimeBlockModel{timeBlock},
				Sun: []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// invalid empty schedule
		{
			schedule: HeatingScheduleResourceModel{},
			expected: false,
		},
		// invalid mixed schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
				MonFri: []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// invalid full schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
				MonFri: []TimeBlockModel{timeBlock},
				Mon:    []TimeBlockModel{timeBlock},
				Tue:    []TimeBlockModel{timeBlock},
				Wed:    []TimeBlockModel{timeBlock},
				Thu:    []TimeBlockModel{timeBlock},
				Fri:    []TimeBlockModel{timeBlock},
				Sat:    []TimeBlockModel{timeBlock},
				Sun:    []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		actual := isMonFriSatSunSchedule(c.schedule)
		if actual != c.expected {
			t.Fatalf("Expected: %t, got: %t", c.expected, actual)
		}
	}
}

func TestIsMonTueWedThuFriSatSunSchedule(t *testing.T) {
	timeBlock := TimeBlockModel{Heating: types.Bool{}, Temperature: types.Float64{}, Start: types.String{}, End: types.String{}}
	cases := []struct {
		schedule HeatingScheduleResourceModel
		expected bool
	}{
		// valid mon-sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// valid mon-fri, sat, sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonFri: []TimeBlockModel{timeBlock},
				Sat:    []TimeBlockModel{timeBlock},
				Sun:    []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// valid mon, tue, wed, thu, fri, sat, sun schedule
		{
			schedule: HeatingScheduleResourceModel{
				Mon: []TimeBlockModel{timeBlock},
				Tue: []TimeBlockModel{timeBlock},
				Wed: []TimeBlockModel{timeBlock},
				Thu: []TimeBlockModel{timeBlock},
				Fri: []TimeBlockModel{timeBlock},
				Sat: []TimeBlockModel{timeBlock},
				Sun: []TimeBlockModel{timeBlock},
			},
			expected: true,
		},
		// invalid empty schedule
		{
			schedule: HeatingScheduleResourceModel{},
			expected: false,
		},
		// invalid mixed schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
				MonFri: []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
		// invalid full schedule
		{
			schedule: HeatingScheduleResourceModel{
				MonSun: []TimeBlockModel{timeBlock},
				MonFri: []TimeBlockModel{timeBlock},
				Mon:    []TimeBlockModel{timeBlock},
				Tue:    []TimeBlockModel{timeBlock},
				Wed:    []TimeBlockModel{timeBlock},
				Thu:    []TimeBlockModel{timeBlock},
				Fri:    []TimeBlockModel{timeBlock},
				Sat:    []TimeBlockModel{timeBlock},
				Sun:    []TimeBlockModel{timeBlock},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		actual := isMonTueWedThuFriSatSunSchedule(c.schedule)
		if actual != c.expected {
			t.Fatalf("Expected: %t, got: %t", c.expected, actual)
		}
	}
}
