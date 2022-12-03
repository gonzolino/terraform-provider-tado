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
var _ datasource.DataSource = &ZoneDataSource{}

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

type ZoneDataSource struct {
	client   *gotado.Tado
	username string
	password string
}

type ZoneDataSourceModel struct {
	ID                         types.Int64  `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Home                       types.String `tfsdk:"home"`
	Type                       types.String `tfsdk:"type"`
	EarlyStart                 types.Bool   `tfsdk:"early_start"`
	DazzleModeEnabled          types.Bool   `tfsdk:"dazzle_mode_enabled"`
	OpenWindowDetectionEnabled types.Bool   `tfsdk:"open_window_detection_enabled"`
}

func (*ZoneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (ZoneDataSource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "A tado zone corresponds to a room in your home. It can contain several tado devices and has its own schedule and configuration.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Zone ID.",
				Type:                types.Int64Type,
				Computed:            true,
			},
			"name": {
				MarkdownDescription: "Name of the zone.",
				Type:                types.StringType,
				Required:            true,
			},
			"home": {
				MarkdownDescription: "The name of the home this zone belongs to.",
				Type:                types.StringType,
				Required:            true,
			},
			"type": {
				MarkdownDescription: "Zone type. Can be either 'Heating' or 'Hot Water'.",
				Type:                types.StringType,
				Computed:            true,
			},
			"early_start": {
				MarkdownDescription: "If true, tado will ensure the desired temperature is already reached when a schedule block starts.",
				Type:                types.BoolType,
				Computed:            true,
			},
			"dazzle_mode_enabled": {
				MarkdownDescription: "If Dazzle Mode is enabled, tado devices in the zone will show an animation when settings are changed via Manual Control.",
				Type:                types.BoolType,
				Computed:            true,
			},
			"open_window_detection_enabled": {
				MarkdownDescription: "If Open Window Detection is enabled, tado devices in the zone will switch off when an open window is detected.",
				Type:                types.BoolType,
				Computed:            true,
			},
		},
	}, nil
}

func (d *ZoneDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ZoneDataSourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := d.client.Me(ctx, d.username, d.password)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	homeName := data.Home.ValueString()
	home, err := me.GetHome(ctx, homeName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", homeName, err))
		return
	}

	zoneName := data.Name.ValueString()
	zone, err := home.GetZone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get zone '%s': %v", zoneName, err))
		return
	}

	earlyStart, err := zone.GetEarlyStart(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable determine if early start is enabled for zone '%s': %v", zone.Name, err))
		return
	}

	data.ID = types.Int64Value(int64(zone.ID))
	data.Name = types.StringValue(zone.Name)
	data.Home = types.StringValue(home.Name)
	data.Type = types.StringValue(string(zone.Type))
	data.EarlyStart = types.BoolValue(earlyStart)
	data.DazzleModeEnabled = types.BoolValue(zone.DazzleMode.Enabled)
	data.OpenWindowDetectionEnabled = types.BoolValue(zone.OpenWindowDetection.Enabled)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
