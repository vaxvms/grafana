package models

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/services/tag"
	"github.com/grafana/grafana/pkg/services/user"
)

type AlertStateType string
type NoDataOption string
type ExecutionErrorOption string

const (
	AlertStateNoData   AlertStateType = "no_data"
	AlertStatePaused   AlertStateType = "paused"
	AlertStateAlerting AlertStateType = "alerting"
	AlertStateOK       AlertStateType = "ok"
	AlertStatePending  AlertStateType = "pending"
	AlertStateUnknown  AlertStateType = "unknown"
)

const (
	NoDataSetOK       NoDataOption = "ok"
	NoDataSetNoData   NoDataOption = "no_data"
	NoDataKeepState   NoDataOption = "keep_state"
	NoDataSetAlerting NoDataOption = "alerting"
)

const (
	ExecutionErrorSetOk       ExecutionErrorOption = "ok"
	ExecutionErrorSetAlerting ExecutionErrorOption = "alerting"
	ExecutionErrorKeepState   ExecutionErrorOption = "keep_state"
)

var (
	ErrCannotChangeStateOnPausedAlert = fmt.Errorf("cannot change state on pause alert")
	ErrRequiresNewState               = fmt.Errorf("update alert state requires a new state")
)

func (s AlertStateType) IsValid() bool {
	return s == AlertStateOK ||
		s == AlertStateNoData ||
		s == AlertStatePaused ||
		s == AlertStatePending ||
		s == AlertStateAlerting ||
		s == AlertStateUnknown
}

func (s NoDataOption) IsValid() bool {
	return s == NoDataSetNoData || s == NoDataSetAlerting || s == NoDataKeepState || s == NoDataSetOK
}

func (s NoDataOption) ToAlertState() AlertStateType {
	return AlertStateType(s)
}

func (s ExecutionErrorOption) IsValid() bool {
	return s == ExecutionErrorSetAlerting || s == ExecutionErrorKeepState || s == ExecutionErrorSetOk
}

func (s ExecutionErrorOption) ToAlertState() AlertStateType {
	return AlertStateType(s)
}

// swagger:model LegacyAlert
type Alert struct {
	Id             int64
	Version        int64
	OrgId          int64
	DashboardId    int64
	PanelId        int64
	Name           string
	Message        string
	Severity       string // Unused
	State          AlertStateType
	Handler        int64 // Unused
	Silenced       bool
	ExecutionError string
	Frequency      int64
	For            time.Duration

	EvalData     *simplejson.Json
	NewStateDate time.Time
	StateChanges int64

	Created time.Time
	Updated time.Time

	Settings *simplejson.Json
}

func (a *Alert) ValidDashboardPanel() bool {
	return a.OrgId != 0 && a.DashboardId != 0 && a.PanelId != 0
}

func (a *Alert) ValidTags() bool {
	for _, tag := range a.GetTagsFromSettings() {
		if len(tag.Key) > 100 || len(tag.Value) > 100 {
			return false
		}
	}
	return true
}

func (a *Alert) ContainsUpdates(other *Alert) bool {
	result := false
	result = result || a.Name != other.Name
	result = result || a.Message != other.Message

	if a.Settings != nil && other.Settings != nil {
		json1, err1 := a.Settings.Encode()
		json2, err2 := other.Settings.Encode()

		if err1 != nil || err2 != nil {
			return false
		}

		result = result || string(json1) != string(json2)
	}

	// don't compare .State! That would be insane.
	return result
}

func (a *Alert) GetTagsFromSettings() []*tag.Tag {
	tags := []*tag.Tag{}
	if a.Settings != nil {
		if data, ok := a.Settings.CheckGet("alertRuleTags"); ok {
			for tagNameString, tagValue := range data.MustMap() {
				// MustMap() already guarantees the return of a `map[string]interface{}`.
				// Therefore we only need to verify that tagValue is a String.
				tagValueString := simplejson.NewFromAny(tagValue).MustString()
				tags = append(tags, &tag.Tag{Key: tagNameString, Value: tagValueString})
			}
		}
	}
	return tags
}

type PauseAlertCommand struct {
	OrgId       int64
	AlertIds    []int64
	ResultCount int64
	Paused      bool
}

type PauseAllAlertCommand struct {
	ResultCount int64
	Paused      bool
}

type SetAlertStateCommand struct {
	AlertId  int64
	OrgId    int64
	State    AlertStateType
	Error    string
	EvalData *simplejson.Json
}

// Queries
type GetAlertsQuery struct {
	OrgId        int64
	State        []string
	DashboardIDs []int64
	PanelId      int64
	Limit        int64
	Query        string
	User         *user.SignedInUser
}

type GetAllAlertsQuery struct{}

type GetAlertByIdQuery struct {
	Id int64
}

type GetAlertStatesForDashboardQuery struct {
	OrgId       int64
	DashboardId int64
}

type AlertListItemDTO struct {
	Id             int64            `json:"id"`
	DashboardId    int64            `json:"dashboardId"`
	DashboardUid   string           `json:"dashboardUid"`
	DashboardSlug  string           `json:"dashboardSlug"`
	PanelId        int64            `json:"panelId"`
	Name           string           `json:"name"`
	State          AlertStateType   `json:"state"`
	NewStateDate   time.Time        `json:"newStateDate"`
	EvalDate       time.Time        `json:"evalDate"`
	EvalData       *simplejson.Json `json:"evalData"`
	ExecutionError string           `json:"executionError"`
	Url            string           `json:"url"`
}

type AlertStateInfoDTO struct {
	Id           int64          `json:"id"`
	DashboardId  int64          `json:"dashboardId"`
	PanelId      int64          `json:"panelId"`
	State        AlertStateType `json:"state"`
	NewStateDate time.Time      `json:"newStateDate"`
}
