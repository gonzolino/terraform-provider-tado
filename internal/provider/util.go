package provider

import (
	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// toTypesString converts a string pointer to a types.String.
// If the pointer is nil, the types.String will be set to Null.
func toTypesString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// boolToPower converts a bool to a gotado.Power.
// If the bool is true, the gotado.Power will be set to On.
// If it is false, it will be set to Off.
func boolToPower(b bool) gotado.Power {
	if b {
		return gotado.PowerOn
	}
	return gotado.PowerOff
}
