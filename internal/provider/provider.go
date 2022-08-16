package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// Use tado client ID and secret from https://app.tado.com/env.js
	tadoClientID     = "tado-web-app"
	tadoClientSecret = "wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &tadoProvider{}

// tadoProvider satisfies the provider.Provider interface and usually is included
// with all Resource and DataSource implementations.
type tadoProvider struct {
	// client can contain the upstream provider SDK or HTTP client used to
	// communicate with the upstream service. Resource and DataSource
	// implementations can then make calls using this client.
	client *gotado.Tado

	username string
	password string

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// providerData can be used to store data from the Terraform configuration.
type providerData struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *tadoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.

	var username string
	if data.Username.Unknown {
		resp.Diagnostics.AddWarning("Tado username is not set", "Tado username is not set. This is required for authentication.")
	}
	if data.Username.Null {
		username = os.Getenv("TADO_USERNAME")
	} else {
		username = data.Username.Value
	}
	if username == "" {
		resp.Diagnostics.AddError("Tado username is not set", "Tado username is not set. This is required for authentication.")
	}

	var password string
	if data.Password.Unknown {
		resp.Diagnostics.AddWarning("Tado password is not set", "Tado password is not set. This is required for authentication.")
	}
	if data.Password.Null {
		password = os.Getenv("TADO_PASSWORD")
	} else {
		password = data.Password.Value
	}
	if password == "" {
		resp.Diagnostics.AddError("Tado password is not set", "Tado password is not set. This is required for authentication.")
	}

	// If the upstream provider SDK or HTTP client requires configuration, such
	// as authentication or logging, this is a great opportunity to do so.
	p.client = gotado.New(tadoClientID, tadoClientSecret)
	p.username = username
	p.password = password

	p.configured = true
}

func (p *tadoProvider) GetResources(_ context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	return map[string]provider.ResourceType{
		"tado_geofencing": geofencingResourceType{},
	}, nil
}

func (p *tadoProvider) GetDataSources(_ context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	return map[string]provider.DataSourceType{
		"tado_home": homeDataSourceType{},
		"tado_zone": zoneDataSourceType{},
	}, nil
}

func (p *tadoProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"username": {
				MarkdownDescription: "Tado username. Can be set via environment variable `TADO_USERNAME`.",
				Optional:            true,
				Type:                types.StringType,
			},
			"password": {
				MarkdownDescription: "Tado Password. Can be set via environment variable `TADO_PASSWORD`.",
				Optional:            true,
				Sensitive:           true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &tadoProvider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in provider.Provider) (tadoProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*tadoProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return tadoProvider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return tadoProvider{}, diags
	}

	return *p, diags
}
