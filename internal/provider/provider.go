package provider

import (
	"context"
	"os"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	// Use tado client ID and secret from https://app.tado.com/env.js
	tadoClientID     = "tado-web-app"
	tadoClientSecret = "wZaRN7rpjn3FoNyF5IFuxg9uMzYJcvOoQ8QWiIqS3hfk6gLhVlG57j5YNoZL2Rtc"
)

// Ensure TadoProvider satisfies various provider interfaces.
var _ provider.Provider = &TadoProvider{}
var _ provider.ProviderWithMetadata = &TadoProvider{}

// TadoProvider defines the provider implementation.
type TadoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TadoProviderModel describes the provider data model.
type TadoProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// tadoProviderData contains data needed to configure tado resources and data
// sources.
type tadoProviderData struct {
	client   *gotado.Tado
	username string
	password string
}

func (p *TadoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tado"
	resp.Version = p.version
}

func (*TadoProvider) Schema(_ context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Tado username. Can be set via environment variable `TADO_USERNAME`.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Tado Password. Can be set via environment variable `TADO_PASSWORD`.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (*TadoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TadoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var username string
	if data.Username.IsUnknown() {
		resp.Diagnostics.AddWarning("Tado username is not set", "Tado username is not set. This is required for authentication.")
	}
	if data.Username.IsNull() {
		username = os.Getenv("TADO_USERNAME")
	} else {
		username = data.Username.ValueString()
	}
	if username == "" {
		resp.Diagnostics.AddError("Tado username is not set", "Tado username is not set. This is required for authentication.")
	}

	var password string
	if data.Password.IsUnknown() {
		resp.Diagnostics.AddWarning("Tado password is not set", "Tado password is not set. This is required for authentication.")
	}
	if data.Password.IsNull() {
		password = os.Getenv("TADO_PASSWORD")
	} else {
		password = data.Password.ValueString()
	}
	if password == "" {
		resp.Diagnostics.AddError("Tado password is not set", "Tado password is not set. This is required for authentication.")
	}

	providerData := &tadoProviderData{
		client:   gotado.New(tadoClientID, tadoClientSecret),
		username: username,
		password: password,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (*TadoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGeofencingResource,
		NewHeatingScheduleResource,
	}
}

func (*TadoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHomeDataSource,
		NewZoneDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TadoProvider{
			version: version,
		}
	}
}
