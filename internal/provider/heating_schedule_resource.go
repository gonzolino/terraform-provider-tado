package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/gonzolino/gotado/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &HeatingScheduleResource{}
var _ resource.ResourceWithImportState = &HeatingScheduleResource{}

func NewHeatingScheduleResource() resource.Resource {
	return &HeatingScheduleResource{}
}

type HeatingScheduleResource struct {
	client *gotado.Tado
}

type TimeBlockModel struct {
	Heating           types.Bool    `tfsdk:"heating"`
	Temperature       types.Float64 `tfsdk:"temperature"`
	Start             types.String  `tfsdk:"start"`
	End               types.String  `tfsdk:"end"`
	GeofencingControl types.Bool    `tfsdk:"geofencing_control"`
}

type HeatingScheduleResourceModel struct {
	ID       types.String     `tfsdk:"id"`
	HomeName types.String     `tfsdk:"home_name"`
	ZoneName types.String     `tfsdk:"zone_name"`
	MonSun   []TimeBlockModel `tfsdk:"mon_sun"`
	MonFri   []TimeBlockModel `tfsdk:"mon_fri"`
	Mon      []TimeBlockModel `tfsdk:"mon"`
	Tue      []TimeBlockModel `tfsdk:"tue"`
	Wed      []TimeBlockModel `tfsdk:"wed"`
	Thu      []TimeBlockModel `tfsdk:"thu"`
	Fri      []TimeBlockModel `tfsdk:"fri"`
	Sat      []TimeBlockModel `tfsdk:"sat"`
	Sun      []TimeBlockModel `tfsdk:"sun"`
}

var timeBlockAttributes = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"heating": schema.BoolAttribute{
			MarkdownDescription: "Whether heating should be turned on or off",
			Required:            true,
		},
		"temperature": schema.Float64Attribute{
			MarkdownDescription: "The temperature to set the heating to. Required when 'heating' is true",
			Optional:            true,
		},
		"start": schema.StringAttribute{
			MarkdownDescription: "When the timeblock starts. Format must be 'hh:mm'.",
			Required:            true,
		},
		"end": schema.StringAttribute{
			MarkdownDescription: "When the timeblock ends. Format must be 'hh:mm'.",
			Required:            true,
		},
		"geofencing_control": schema.BoolAttribute{
			MarkdownDescription: "Whether the settings of this time block are overwritten by the tado away settings. Defaults to 'true'.",
			Optional:            true,
			Computed:            true,
		},
	},
}

func (*HeatingScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_heating_schedule"
}

func (HeatingScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The heating schedule of a zone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of this heating schedule resource.",
				Computed:            true,
			},
			"home_name": schema.StringAttribute{
				MarkdownDescription: "Name of the home this heating schedule resource belongs to.",
				Required:            true,
			},
			"zone_name": schema.StringAttribute{
				MarkdownDescription: "Name of the zone of this heating schedule.",
				Required:            true,
			},
			"mon_sun": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Monday - Sunday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"mon_fri": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Monday - Friday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"mon": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Monday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"tue": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Tuesday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"wed": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Wednesday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"thu": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Thursday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"fri": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Friday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"sat": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Saturday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
			"sun": schema.ListNestedAttribute{
				MarkdownDescription: "Schedule for Sunday.",
				Optional:            true,
				NestedObject:        timeBlockAttributes,
			},
		},
	}
}

func (r *HeatingScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = gotado.NewWithTokenRefreshCallback(ctx, data.config, data.token, createTokenUpdateCallback(data.tokenPath, &resp.Diagnostics))
}

func (r HeatingScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HeatingScheduleResourceModel

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

	zoneName := data.ZoneName.ValueString()
	zone, err := home.GetZone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get zone '%s': %v", zoneName, err))
		return
	}

	schedule, diags := heatingScheduleResourceModelToObject(ctx, data, zone)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := zone.SetHeatingSchedule(ctx, schedule); err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to create heating schedule for zone '%s': %v", zone.Name, err))
		return
	}

	schedule, err = zone.GetHeatingSchedule(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get created heating schedule for zone '%s': %v", zone.Name, err))
		return
	}

	heatingScheduleToResourceData(ctx, schedule, &data)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HeatingScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HeatingScheduleResourceModel

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

	zoneName := data.ZoneName.ValueString()
	zone, err := home.GetZone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get zone '%s': %v", zoneName, err))
		return
	}

	schedule, err := zone.GetHeatingSchedule(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get heating schedule for zone '%s': %v", zone.Name, err))
		return
	}

	heatingScheduleToResourceData(ctx, schedule, &data)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r HeatingScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HeatingScheduleResourceModel

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

	zoneName := data.ZoneName.ValueString()
	zone, err := home.GetZone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get zone '%s': %v", zoneName, err))
		return
	}

	schedule, diags := heatingScheduleResourceModelToObject(ctx, data, zone)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := zone.SetHeatingSchedule(ctx, schedule); err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to create heating schedule for zone '%s': %v", zone.Name, err))
		return
	}

	schedule, err = zone.GetHeatingSchedule(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Tado API Error", fmt.Sprintf("Unable to get created heating schedule for zone '%s': %v", zone.Name, err))
		return
	}

	heatingScheduleToResourceData(ctx, schedule, &data)
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (HeatingScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HeatingScheduleResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// A schedule can't be deleted, so we simply 'forget' it
}

func (HeatingScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	splittedID := strings.Split(req.ID, "/")
	if len(splittedID) != 2 {
		resp.Diagnostics.AddError("Resource Import ID invalid", fmt.Sprintf("ID '%s' should be in format 'home_name/zone_name'", req.ID))
		return
	}

	homeName, zoneName := splittedID[0], splittedID[1]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("home_name"), homeName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_name"), zoneName)...)
}

// isMonSunSchedule checks if the heating schedule has a valid Monday - Sunday schedule
func isMonSunSchedule(data HeatingScheduleResourceModel) bool {
	return data.MonSun != nil && data.MonFri == nil && data.Mon == nil && data.Tue == nil && data.Wed == nil && data.Thu == nil && data.Fri == nil && data.Sat == nil && data.Sun == nil
}

// isMonFriSatSunSchedule checks if the heating schedule has a valid Monday - Friday, Saturday, Sunday schedule
func isMonFriSatSunSchedule(data HeatingScheduleResourceModel) bool {
	return data.MonSun == nil && data.MonFri != nil && data.Mon == nil && data.Tue == nil && data.Wed == nil && data.Thu == nil && data.Fri == nil && data.Sat != nil && data.Sun != nil
}

// isMonTueWedThuFriSatSunSchedule checks if the heating schedule has a valid Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday schedule
func isMonTueWedThuFriSatSunSchedule(data HeatingScheduleResourceModel) bool {
	return data.MonSun == nil && data.MonFri == nil && data.Mon != nil && data.Tue != nil && data.Wed != nil && data.Thu != nil && data.Fri != nil && data.Sat != nil && data.Sun != nil
}

func heatingScheduleToResourceData(ctx context.Context, schedule *gotado.HeatingSchedule, data *HeatingScheduleResourceModel) {
	homeName, zoneName := data.HomeName.ValueString(), data.ZoneName.ValueString()
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", homeName, zoneName))
	data.HomeName = types.StringValue(homeName)
	data.ZoneName = types.StringValue(zoneName)

	sortedBlocks := sortTimeBlocksByDayType(schedule.Blocks)

	switch schedule.ScheduleDays {
	case gotado.ScheduleDaysMonToSun:
		data.MonSun = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeMondayToSunday]))
		for i, block := range sortedBlocks[gotado.DayTypeMondayToSunday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.MonSun[i])
		}
	case gotado.ScheduleDaysMonToFriSatSun:
		data.MonFri = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeMondayToFriday]))
		data.Sat = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeSaturday]))
		data.Sun = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeSunday]))
		for i, block := range sortedBlocks[gotado.DayTypeMondayToFriday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.MonFri[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeSaturday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Sat[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeSunday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Sun[i])
		}
	case gotado.ScheduleDaysMonTueWedThuFriSatSun:
		data.Mon = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeMonday]))
		data.Tue = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeTuesday]))
		data.Wed = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeWednesday]))
		data.Thu = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeThursday]))
		data.Fri = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeFriday]))
		data.Sat = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeSaturday]))
		data.Sun = make([]TimeBlockModel, len(sortedBlocks[gotado.DayTypeSunday]))
		for i, block := range sortedBlocks[gotado.DayTypeMonday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Mon[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeTuesday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Tue[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeWednesday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Wed[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeThursday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Thu[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeFriday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Fri[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeSaturday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Sat[i])
		}
		for i, block := range sortedBlocks[gotado.DayTypeSunday] {
			timeBlockObjectToTimeBlockModel(ctx, block, &data.Sun[i])
		}
	}
}

func heatingScheduleResourceModelToObject(ctx context.Context, data HeatingScheduleResourceModel, zone *gotado.Zone) (*gotado.HeatingSchedule, diag.Diagnostics) {
	var err error
	var schedule *gotado.HeatingSchedule
	diags := diag.Diagnostics{}
	switch {
	case isMonSunSchedule(data):
		schedule, err = zone.ScheduleMonToSun(ctx)
		if err != nil {
			diags.AddError("Tado API Error", fmt.Sprintf("Unable to initialize schedule for zone '%s': %v", zone.Name, err))
			return nil, diags
		}
		first := data.MonSun[0]
		power := boolToPower(first.Heating.ValueBool())
		schedule.NewTimeBlock(ctx, gotado.DayTypeMondayToSunday,
			first.Start.ValueString(),
			first.End.ValueString(),
			first.GeofencingControl.ValueBool(),
			power,
			first.Temperature.ValueFloat64())
		for _, block := range data.MonSun[1:] {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeMondayToSunday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
	case isMonFriSatSunSchedule(data):
		schedule, err = zone.ScheduleMonToFriSatSun(ctx)
		if err != nil {
			diags.AddError("Tado API Error", fmt.Sprintf("Unable to initialize schedule for zone '%s': %v", zone.Name, err))
			return nil, diags
		}
		first := data.MonFri[0]
		power := boolToPower(first.Heating.ValueBool())
		schedule.NewTimeBlock(ctx, gotado.DayTypeMondayToFriday,
			first.Start.ValueString(),
			first.End.ValueString(),
			first.GeofencingControl.ValueBool(),
			power,
			first.Temperature.ValueFloat64())
		for _, block := range data.MonFri[1:] {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeMondayToFriday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Sat {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeSaturday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Sun {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeSunday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
	case isMonTueWedThuFriSatSunSchedule(data):
		schedule, err = zone.ScheduleAllDays(ctx)
		if err != nil {
			diags.AddError("Tado API Error", fmt.Sprintf("Unable to initialize schedule for zone '%s': %v", zone.Name, err))
			return nil, diags
		}
		first := data.Mon[0]
		power := boolToPower(first.Heating.ValueBool())
		schedule.NewTimeBlock(ctx, gotado.DayTypeMonday,
			first.Start.ValueString(),
			first.End.ValueString(),
			first.GeofencingControl.ValueBool(),
			power,
			first.Temperature.ValueFloat64())
		for _, block := range data.Mon[1:] {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeMonday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Tue {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeTuesday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Wed {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeWednesday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Thu {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeThursday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Fri {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeFriday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Sat {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeSaturday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
		for _, block := range data.Sun {
			power := boolToPower(block.Heating.ValueBool())
			schedule.AddTimeBlock(ctx, gotado.DayTypeSunday,
				block.Start.ValueString(),
				block.End.ValueString(),
				block.GeofencingControl.ValueBool(),
				power,
				block.Temperature.ValueFloat64())
		}
	default:
		diags.AddError("Invalid Heating Schedule", fmt.Sprintf("Unable to create heating schedule for zone '%s': No valid schedule provided", zone.Name))
		return nil, diags
	}
	return schedule, nil
}

func sortTimeBlocksByDayType(blocks []*gotado.ScheduleTimeBlock) map[gotado.DayType][]*gotado.ScheduleTimeBlock {
	sortedBlocks := make(map[gotado.DayType][]*gotado.ScheduleTimeBlock, len(blocks))

	for _, block := range blocks {
		if _, ok := sortedBlocks[block.DayType]; !ok {
			sortedBlocks[block.DayType] = make([]*gotado.ScheduleTimeBlock, 0)
		}
		sortedBlocks[block.DayType] = append(sortedBlocks[block.DayType], block)
	}

	return sortedBlocks
}

func timeBlockObjectToTimeBlockModel(_ context.Context, block *gotado.ScheduleTimeBlock, model *TimeBlockModel) {
	model.Heating = types.BoolValue(block.Setting.Power == "ON")
	if block.Setting.Temperature != nil {
		model.Temperature = types.Float64Value(block.Setting.Temperature.Celsius)
	}
	model.Start = types.StringValue(block.Start)
	model.End = types.StringValue(block.End)
	model.GeofencingControl = types.BoolValue(!block.GeolocationOverride)
}
