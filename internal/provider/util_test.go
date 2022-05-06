package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToTypesStr(t *testing.T) {
	cases := []struct {
		s        string
		expected types.String
	}{
		{
			s:        "test",
			expected: types.String{Value: "test"},
		},
	}

	for _, c := range cases {
		actual := toTypesString(&c.s)
		if !actual.Equal(c.expected) {
			t.Fatalf("Expected: %#v, got: %#v", c.expected, actual)
		}
	}
}
