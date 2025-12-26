package provider

import (
	"context"
	"fmt"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &HomeDataSource{}

func NewHomeDataSource() datasource.DataSource {
	return &HomeDataSource{}
}

type HomeDataSource struct {
	client *gotado.Tado
}

type HomeDataSourceModel struct {
	ID              types.Int64   `tfsdk:"id"`
	Name            types.String  `tfsdk:"name"`
	TemperatureUnit types.String  `tfsdk:"temperature_unit"`
	ContactName     types.String  `tfsdk:"contact_name"`
	ContactEmail    types.String  `tfsdk:"contact_email"`
	ContactPhone    types.String  `tfsdk:"contact_phone"`
	AddressLine1    types.String  `tfsdk:"address_line1"`
	AddressLine2    types.String  `tfsdk:"address_line2"`
	AddressZipcode  types.String  `tfsdk:"address_zipcode"`
	AddressCity     types.String  `tfsdk:"address_city"`
	AddressState    types.String  `tfsdk:"address_state"`
	AddressCountry  types.String  `tfsdk:"address_country"`
	GeolocationLat  types.Float64 `tfsdk:"geolocation_lat"`
	GeolocationLong types.Float64 `tfsdk:"geolocation_long"`
}

func (*HomeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_home"
}

func (HomeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A tado home holds all tado devices and heating zones. The home data source provides information such as contact details, address, etc.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Home ID.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the home.",
				Required:            true,
			},
			"temperature_unit": schema.StringAttribute{
				MarkdownDescription: "Temperature unit used in the home. Either 'Celsius' or 'Fahrenheit'.",
				Computed:            true,
			},
			"contact_name": schema.StringAttribute{
				MarkdownDescription: "Name of the contact person.",
				Computed:            true,
			},
			"contact_email": schema.StringAttribute{
				MarkdownDescription: "Email address of the contact person.",
				Computed:            true,
			},
			"contact_phone": schema.StringAttribute{
				MarkdownDescription: "Phone number of the contact person.",
				Computed:            true,
			},
			"address_line1": schema.StringAttribute{
				MarkdownDescription: "Address line 1.",
				Computed:            true,
			},
			"address_line2": schema.StringAttribute{
				MarkdownDescription: "Address line 2.",
				Computed:            true,
			},
			"address_zipcode": schema.StringAttribute{
				MarkdownDescription: "Zip code.",
				Computed:            true,
			},
			"address_city": schema.StringAttribute{
				MarkdownDescription: "City.",
				Computed:            true,
			},
			"address_state": schema.StringAttribute{
				MarkdownDescription: "State.",
				Computed:            true,
			},
			"address_country": schema.StringAttribute{
				MarkdownDescription: "Country.",
				Computed:            true,
			},
			"geolocation_lat": schema.Float64Attribute{
				MarkdownDescription: "Latitude used for Geofencing.",
				Computed:            true,
			},
			"geolocation_long": schema.Float64Attribute{
				MarkdownDescription: "Longitude used for Geofencing.",
				Computed:            true,
			},
		},
	}
}

func (d *HomeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*tadoProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *tadoProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = gotado.NewWithTokenRefreshCallback(ctx, data.config, data.token, createTokenUpdateCallback(data.token_path, &resp.Diagnostics))
}

func (d HomeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HomeDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := d.client.Me(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado Authentication Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	home, err := me.GetHome(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Tado Authentication Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	data.ID = types.Int64Value(int64(home.ID))
	data.Name = types.StringValue(home.Name)
	data.TemperatureUnit = types.StringValue(string(home.TemperatureUnit))
	data.ContactName = toTypesString(home.ContactDetails.Name)
	data.ContactEmail = toTypesString(home.ContactDetails.Email)
	data.ContactPhone = toTypesString(home.ContactDetails.Phone)
	data.AddressLine1 = toTypesString(home.Address.AddressLine1)
	data.AddressLine2 = toTypesString(home.Address.AddressLine2)
	data.AddressZipcode = toTypesString(home.Address.ZipCode)
	data.AddressCity = toTypesString(home.Address.City)
	data.AddressState = toTypesString(home.Address.State)
	data.AddressCountry = toTypesString(home.Address.Country)
	data.GeolocationLat = types.Float64Value(home.Geolocation.Latitude)
	data.GeolocationLong = types.Float64Value(home.Geolocation.Longitude)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
