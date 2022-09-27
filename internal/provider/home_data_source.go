package provider

import (
	"context"
	"fmt"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &HomeDataSource{}

func NewHomeDataSource() datasource.DataSource {
	return &HomeDataSource{}
}

type HomeDataSource struct {
	client   *gotado.Tado
	username string
	password string
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

func (HomeDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "A tado home holds all tado devices and heating zones. The home data source provides information such as contact details, address, etc.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Home ID.",
				Type:                types.Int64Type,
				Computed:            true,
			},
			"name": {
				MarkdownDescription: "Name of the home.",
				Type:                types.StringType,
				Required:            true,
			},
			"temperature_unit": {
				MarkdownDescription: "Temperature unit used in the home. Either 'Celsius' or 'Fahrenheit'.",
				Type:                types.StringType,
				Computed:            true,
			},
			"contact_name": {
				MarkdownDescription: "Name of the contact person.",
				Type:                types.StringType,
				Computed:            true,
			},
			"contact_email": {
				MarkdownDescription: "Email address of the contact person.",
				Type:                types.StringType,
				Computed:            true,
			},
			"contact_phone": {
				MarkdownDescription: "Phone number of the contact person.",
				Type:                types.StringType,
				Computed:            true,
			},
			"address_line1": {
				MarkdownDescription: "Address line 1.",
				Type:                types.StringType,
				Computed:            true,
			},
			"address_line2": {
				MarkdownDescription: "Address line 2.",
				Type:                types.StringType,
				Computed:            true,
			},
			"address_zipcode": {
				MarkdownDescription: "Zip code.",
				Type:                types.StringType,
				Computed:            true,
			},
			"address_city": {
				MarkdownDescription: "City.",
				Type:                types.StringType,
				Computed:            true,
			},
			"address_state": {
				MarkdownDescription: "State.",
				Type:                types.StringType,
				Computed:            true,
			},
			"address_country": {
				MarkdownDescription: "Country.",
				Type:                types.StringType,
				Computed:            true,
			},
			"geolocation_lat": {
				MarkdownDescription: "Latitude used for Geofencing.",
				Type:                types.Float64Type,
				Computed:            true,
			},
			"geolocation_long": {
				MarkdownDescription: "Longitude used for Geofencing.",
				Type:                types.Float64Type,
				Computed:            true,
			},
		},
	}, nil
}

func (d *HomeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = data.client
	d.username = data.username
	d.password = data.password
}

func (d HomeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HomeDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := d.client.Me(ctx, d.username, d.password)
	if err != nil {
		resp.Diagnostics.AddError("Tado Authentication Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	home, err := me.GetHome(ctx, data.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("Tado Authentication Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	data.ID = types.Int64{Value: int64(home.ID)}
	data.Name = types.String{Value: home.Name}
	data.TemperatureUnit = types.String{Value: string(home.TemperatureUnit)}
	data.ContactName = toTypesString(home.ContactDetails.Name)
	data.ContactEmail = toTypesString(home.ContactDetails.Email)
	data.ContactPhone = toTypesString(home.ContactDetails.Phone)
	data.AddressLine1 = toTypesString(home.Address.AddressLine1)
	data.AddressLine2 = toTypesString(home.Address.AddressLine2)
	data.AddressZipcode = toTypesString(home.Address.ZipCode)
	data.AddressCity = toTypesString(home.Address.City)
	data.AddressState = toTypesString(home.Address.State)
	data.AddressCountry = toTypesString(home.Address.Country)
	data.GeolocationLat = types.Float64{Value: home.Geolocation.Latitude}
	data.GeolocationLong = types.Float64{Value: home.Geolocation.Longitude}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
