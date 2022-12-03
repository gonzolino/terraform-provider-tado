package provider

import (
	"testing"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToTypesStr(t *testing.T) {
	cases := []struct {
		s        string
		expected types.String
	}{
		{
			s:        "test",
			expected: types.StringValue("test"),
		},
	}

	for _, c := range cases {
		actual := toTypesString(&c.s)
		if !actual.Equal(c.expected) {
			t.Fatalf("Expected: %#v, got: %#v", c.expected, actual)
		}
	}
}

func TestBoolToPower(t *testing.T) {
	if boolToPower(true) != gotado.PowerOn {
		t.Fatalf("Expected: %s, got: %s", gotado.PowerOn, boolToPower(true))
	}

	if boolToPower(false) != gotado.PowerOff {
		t.Fatalf("Expected: %s, got: %s", gotado.PowerOff, boolToPower(false))
	}
}
