package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.ResourceType = geofencingResourceType{}
var _ resource.Resource = geofencingResource{}
var _ resource.ResourceWithImportState = geofencingResource{}

type geofencingResourceType struct{}

func (geofencingResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Controls geofencing of a home.",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "ID of this geofencing resource. This should match the home_name.",
				Type:                types.StringType,
				Computed:            true,
			},
			"home_name": {
				MarkdownDescription: "Name of the home this geofencing resource belongs to.",
				Type:                types.StringType,
				Required:            true,
			},
			"presence": {
				MarkdownDescription: "Whether somebody is present in the home. Can be one of 'auto', 'home' or 'away'.",
				Type:                types.StringType,
				Required:            true,
			},
		},
	}, nil
}

func (geofencingResourceType) NewResource(_ context.Context, in provider.Provider) (resource.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return geofencingResource{
		provider: provider,
	}, diags
}

type geofencingResourceData struct {
	ID       types.String `tfsdk:"id"`
	HomeName types.String `tfsdk:"home_name"`
	Presence types.String `tfsdk:"presence"`
}

type geofencingResource struct {
	provider tadoProvider
}

func (r geofencingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data geofencingResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := r.provider.client.Me(ctx, r.provider.username, r.provider.password)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	home, err := me.GetHome(ctx, data.HomeName.Value)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", data.HomeName.Value, err))
		return
	}

	switch data.Presence.Value {
	case "auto":
		err = home.SetPresenceAuto(ctx)
	case "home":
		err = home.SetPresenceHome(ctx)
	case "away":
		err = home.SetPresenceAway(ctx)
	default:
		resp.Diagnostics.AddError("Invalid Presence", fmt.Sprintf("Invalid presence value '%s', must be one of 'auto', 'home' or 'away'.", data.Presence.Value))
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to set presence to '%s': %v", data.Presence.Value, err))
		return
	}

	homeState, err := home.GetState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get state of home '%s': %v", data.HomeName.Value, err))
		return
	}

	presence := strings.ToLower(string(homeState.Presence))
	// If presence is not locked, it is set to 'auto'.
	if !homeState.PresenceLocked {
		presence = "auto"
	}

	data.ID = types.String{Value: home.Name}
	data.HomeName = types.String{Value: home.Name}
	data.Presence = types.String{Value: presence}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r geofencingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data geofencingResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := r.provider.client.Me(ctx, r.provider.username, r.provider.password)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	home, err := me.GetHome(ctx, data.HomeName.Value)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", data.HomeName.Value, err))
		return
	}

	homeState, err := home.GetState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get state of home '%s': %v", data.HomeName.Value, err))
		return
	}

	presence := strings.ToLower(string(homeState.Presence))
	// If presence is not locked, it is set to 'auto'.
	if !homeState.PresenceLocked {
		presence = "auto"
	}

	data.ID = types.String{Value: home.Name}
	data.HomeName = types.String{Value: home.Name}
	data.Presence = types.String{Value: presence}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r geofencingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data geofencingResourceData

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := r.provider.client.Me(ctx, r.provider.username, r.provider.password)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	home, err := me.GetHome(ctx, data.HomeName.Value)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", data.HomeName.Value, err))
		return
	}

	switch data.Presence.Value {
	case "auto":
		err = home.SetPresenceAuto(ctx)
	case "home":
		err = home.SetPresenceHome(ctx)
	case "away":
		err = home.SetPresenceAway(ctx)
	default:
		resp.Diagnostics.AddError("Invalid Presence", fmt.Sprintf("Invalid presence value '%s', must be one of 'auto', 'home' or 'away'.", data.Presence.Value))
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to set presence to '%s': %v", data.Presence.Value, err))
		return
	}

	homeState, err := home.GetState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get state of home '%s': %v", data.HomeName.Value, err))
		return
	}

	presence := strings.ToLower(string(homeState.Presence))
	// If presence is not locked, it is set to 'auto'.
	if !homeState.PresenceLocked {
		presence = "auto"
	}

	data.ID = types.String{Value: home.Name}
	data.HomeName = types.String{Value: home.Name}
	data.Presence = types.String{Value: presence}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (geofencingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data geofencingResourceData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// No deletion necesary on tado api.
}

func (geofencingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("home_name"), req, resp)
}
