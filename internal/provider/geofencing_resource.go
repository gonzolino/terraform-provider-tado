package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &GeofencingResource{}
var _ resource.ResourceWithImportState = &GeofencingResource{}

func NewGeofencingResource() resource.Resource {
	return &GeofencingResource{}
}

type GeofencingResource struct {
	client *gotado.Tado
}

type GeofencingResourceModel struct {
	ID       types.String `tfsdk:"id"`
	HomeName types.String `tfsdk:"home_name"`
	Presence types.String `tfsdk:"presence"`
}

func (*GeofencingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_geofencing"
}

func (GeofencingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Controls geofencing of a home.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of this geofencing resource. This should match the home_name.",
				Computed:            true,
			},
			"home_name": schema.StringAttribute{
				MarkdownDescription: "Name of the home this geofencing resource belongs to.",
				Required:            true,
			},
			"presence": schema.StringAttribute{
				MarkdownDescription: "Whether somebody is present in the home. Can be one of 'auto', 'home' or 'away'.",
				Required:            true,
			},
		},
	}
}

func (r *GeofencingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*tadoProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tadoProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = gotado.NewWithTokenRefreshCallback(ctx, data.config, data.token, createTokenUpdateCallback(data.token_path, &resp.Diagnostics))
}

func (r GeofencingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GeofencingResourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := r.client.Me(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	homeName := data.HomeName.ValueString()
	home, err := me.GetHome(ctx, homeName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", homeName, err))
		return
	}

	presence := data.Presence.ValueString()
	switch presence {
	case "auto":
		err = home.SetPresenceAuto(ctx)
	case "home":
		err = home.SetPresenceHome(ctx)
	case "away":
		err = home.SetPresenceAway(ctx)
	default:
		resp.Diagnostics.AddError("Invalid Presence", fmt.Sprintf("Invalid presence value '%s', must be one of 'auto', 'home' or 'away'.", presence))
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to set presence to '%s': %v", presence, err))
		return
	}

	homeState, err := home.GetState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get state of home '%s': %v", homeName, err))
		return
	}

	presence = strings.ToLower(string(homeState.Presence))
	// If presence is not locked, it is set to 'auto'.
	if !homeState.PresenceLocked {
		presence = "auto"
	}

	data.ID = types.StringValue(home.Name)
	data.HomeName = types.StringValue(home.Name)
	data.Presence = types.StringValue(presence)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r GeofencingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GeofencingResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := r.client.Me(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	homeName := data.HomeName.ValueString()
	home, err := me.GetHome(ctx, homeName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", homeName, err))
		return
	}

	homeState, err := home.GetState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get state of home '%s': %v", homeName, err))
		return
	}

	presence := strings.ToLower(string(homeState.Presence))
	// If presence is not locked, it is set to 'auto'.
	if !homeState.PresenceLocked {
		presence = "auto"
	}

	data.ID = types.StringValue(home.Name)
	data.HomeName = types.StringValue(home.Name)
	data.Presence = types.StringValue(presence)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r GeofencingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GeofencingResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	me, err := r.client.Me(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to authenticate with Tado: %v", err))
		return
	}

	homeName := data.HomeName.ValueString()
	home, err := me.GetHome(ctx, homeName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get home '%s': %v", homeName, err))
		return
	}

	presence := data.Presence.ValueString()
	switch presence {
	case "auto":
		err = home.SetPresenceAuto(ctx)
	case "home":
		err = home.SetPresenceHome(ctx)
	case "away":
		err = home.SetPresenceAway(ctx)
	default:
		resp.Diagnostics.AddError("Invalid Presence", fmt.Sprintf("Invalid presence value '%s', must be one of 'auto', 'home' or 'away'.", presence))
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to set presence to '%s': %v", presence, err))
		return
	}

	homeState, err := home.GetState(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get state of home '%s': %v", homeName, err))
		return
	}

	presence = strings.ToLower(string(homeState.Presence))
	// If presence is not locked, it is set to 'auto'.
	if !homeState.PresenceLocked {
		presence = "auto"
	}

	data.ID = types.StringValue(home.Name)
	data.HomeName = types.StringValue(home.Name)
	data.Presence = types.StringValue(presence)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (GeofencingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GeofencingResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// No deletion necesary on tado api.
}

func (GeofencingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("home_name"), req, resp)
}
