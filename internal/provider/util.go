package provider

import (
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
