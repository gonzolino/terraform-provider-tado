package provider

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/oauth2"
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

func updateToken(token *oauth2.Token, path string) error {
	tokenBytes, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(tokenBytes)
	return err
}

func readToken(path string) (*oauth2.Token, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(raw, &token); err != nil {
		return nil, fmt.Errorf("unable to unmarshal token: %w", err)
	}

	return &token, nil
}

func createTokenUpdateCallback(token_path string, diagnostics *diag.Diagnostics) func(token *oauth2.Token) {
	return func(token *oauth2.Token) {
		if err := updateToken(token, token_path); err != nil {
			diagnostics.AddWarning(
				"Unable to update token",
				fmt.Sprintf("Failed to update token at %s: %v", token_path, err),
			)
		}
	}
}
