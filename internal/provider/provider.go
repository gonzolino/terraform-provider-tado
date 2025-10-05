package provider

import (
	"context"
	"fmt"

	"github.com/cli/browser"
	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	tadoClientID = "1bb50063-6b0c-4d11-bd99-387f4a91cc46"
)

// Ensure TadoProvider satisfies various provider interfaces.
var _ provider.Provider = &TadoProvider{}

// TadoProvider defines the provider implementation.
type TadoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TadoProviderModel describes the provider data model.
type TadoProviderModel struct {
}

// tadoProviderData contains data needed to configure tado resources and data
// sources.
type tadoProviderData struct {
	client *gotado.Tado
}

func (p *TadoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "tado"
	resp.Version = p.version
}

func (*TadoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The Tado provider is used to manage your Tado home.

While not everything is supported yet, the provider is able to manage heating schedules and settings such as geofencing.
`,
	}
}

func (*TadoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TadoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	config := gotado.AuthConfig(tadoClientID)

	deviceAuth, err := config.DeviceAuth(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to start authentication",
			fmt.Sprintf("Failed to initiate device authentication: %v", err),
		)
		return
	}

	if err := browser.OpenURL(deviceAuth.VerificationURIComplete); err != nil {
		resp.Diagnostics.AddWarning(
			"Unable to open browser",
			fmt.Sprintf("Please visit %s to authenticate: %v", deviceAuth.VerificationURIComplete, err),
		)
	}

	token, err := config.DeviceAccessToken(ctx, deviceAuth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Authentication failed",
			fmt.Sprintf("Failed to authenticate with Tado: %v", err),
		)
		return
	}

	providerData := &tadoProviderData{
		client: gotado.New(ctx, config, token),
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
