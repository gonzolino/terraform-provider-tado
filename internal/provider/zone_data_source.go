package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = zoneDataSourceType{}
var _ tfsdk.DataSource = zoneDataSource{}

type zoneDataSourceType struct{}

func (zoneDataSourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

func (zoneDataSourceType) NewDataSource(_ context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return zoneDataSource{
		provider: provider,
	}, diags
}

type zoneDataSourceData struct {
	ID                         types.Int64  `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Home                       types.String `tfsdk:"home"`
	Type                       types.String `tfsdk:"type"`
	EarlyStart                 types.Bool   `tfsdk:"early_start"`
	DazzleModeEnabled          types.Bool   `tfsdk:"dazzle_mode_enabled"`
	OpenWindowDetectionEnabled types.Bool   `tfsdk:"open_window_detection_enabled"`
}

type zoneDataSource struct {
	provider provider
}

func (d zoneDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data zoneDataSourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := d.provider.client.Me(ctx, d.provider.username, d.provider.password)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	home, err := me.GetHome(ctx, data.Home.Value)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", data.Home.Value, err))
		return
	}

	zone, err := home.GetZone(ctx, data.Name.Value)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get zone '%s': %v", data.Name.Value, err))
		return
	}

	earlyStart, err := zone.GetEarlyStart(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable determine if early start is enabled for zone '%s': %v", zone.Name, err))
		return
	}

	data.ID = types.Int64{Value: int64(zone.ID)}
	data.Name = types.String{Value: zone.Name}
	data.Home = types.String{Value: home.Name}
	data.Type = types.String{Value: string(zone.Type)}
	data.EarlyStart = types.Bool{Value: earlyStart}
	data.DazzleModeEnabled = types.Bool{Value: zone.DazzleMode.Enabled}
	data.OpenWindowDetectionEnabled = types.Bool{Value: zone.OpenWindowDetection.Enabled}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
