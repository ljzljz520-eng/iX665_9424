package pumpstation

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StatusRunning = "运行"
	StatusStopped = "停运"
	StatusRepair  = "维修"
	StatusRetired = "报废"
)

var allowedStatuses = map[string]struct{}{
	StatusRunning: {},
	StatusStopped: {},
	StatusRepair:  {},
	StatusRetired: {},
}

type Equipment struct {
	ID               string       `json:"id"`
	StationName      string       `json:"station_name"`
	EquipmentNumber  string       `json:"equipment_number"`
	Power            float64      `json:"power"`
	InstallationSite string       `json:"installation_site"`
	ResponsibleTeam  string       `json:"responsible_team"`
	CommissionedOn   string       `json:"commissioned_on"`
	RunningStatus    string       `json:"running_status"`
	Attachments      []Attachment `json:"attachments"`
}

type Attachment struct {
	ID         string `json:"id"`
	FileName   string `json:"file_name"`
	StorageKey string `json:"storage_key"`
	Size       int64  `json:"size"`
}

type OperationLog struct {
	ID          int64  `json:"id"`
	EquipmentID string `json:"equipment_id"`
	Action      string `json:"action"`
	Detail      string `json:"detail"`
	Result      string `json:"result"`
}

type CreateRequest struct {
	StationName      string       `json:"station_name"`
	EquipmentNumber  string       `json:"equipment_number"`
	Power            float64      `json:"power"`
	InstallationSite string       `json:"installation_site"`
	ResponsibleTeam  string       `json:"responsible_team"`
	CommissionedOn   string       `json:"commissioned_on"`
	RunningStatus    string       `json:"running_status"`
	Attachments      []Attachment `json:"attachments"`
}

type EditRequest struct {
	StationName      string       `json:"station_name"`
	EquipmentNumber  string       `json:"equipment_number"`
	Power            float64      `json:"power"`
	InstallationSite string       `json:"installation_site"`
	ResponsibleTeam  string       `json:"responsible_team"`
	CommissionedOn   string       `json:"commissioned_on"`
	RunningStatus    string       `json:"running_status"`
	Attachments      []Attachment `json:"attachments"`
}

type Page struct {
	Items      []Equipment `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

func (r CreateRequest) equipment(id string) Equipment {
	return Equipment{
		ID:               id,
		StationName:      strings.TrimSpace(r.StationName),
		EquipmentNumber:  strings.TrimSpace(r.EquipmentNumber),
		Power:            r.Power,
		InstallationSite: strings.TrimSpace(r.InstallationSite),
		ResponsibleTeam:  strings.TrimSpace(r.ResponsibleTeam),
		CommissionedOn:   strings.TrimSpace(r.CommissionedOn),
		RunningStatus:    strings.TrimSpace(r.RunningStatus),
		Attachments:      cloneAttachments(r.Attachments),
	}
}

func (r EditRequest) equipment(id string) Equipment {
	return CreateRequest{
		StationName:      r.StationName,
		EquipmentNumber:  r.EquipmentNumber,
		Power:            r.Power,
		InstallationSite: r.InstallationSite,
		ResponsibleTeam:  r.ResponsibleTeam,
		CommissionedOn:   r.CommissionedOn,
		RunningStatus:    r.RunningStatus,
		Attachments:      r.Attachments,
	}.equipment(id)
}

func validateEquipment(e Equipment) error {
	fields := []struct {
		name  string
		value string
	}{
		{"泵站名称", e.StationName},
		{"设备编号", e.EquipmentNumber},
		{"安装位置", e.InstallationSite},
		{"责任班组", e.ResponsibleTeam},
		{"投运日期", e.CommissionedOn},
	}
	for _, field := range fields {
		if field.value == "" {
			return fmt.Errorf("%s不能为空", field.name)
		}
	}
	if e.Power <= 0 {
		return fmt.Errorf("功率必须大于0")
	}
	if _, ok := allowedStatuses[e.RunningStatus]; !ok {
		return fmt.Errorf("运行状态不受支持: %s", e.RunningStatus)
	}
	if _, err := time.Parse("2006-01-02", e.CommissionedOn); err != nil {
		return fmt.Errorf("投运日期格式无效")
	}
	if err := validateAttachments(e.Attachments); err != nil {
		return err
	}
	return nil
}

func validateAttachments(items []Attachment) error {
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, err := uuid.Parse(item.ID); err != nil {
			return fmt.Errorf("附件编号无效")
		}
		if item.FileName == "" || item.StorageKey == "" {
			return fmt.Errorf("附件信息不完整")
		}
		if item.Size < 0 {
			return fmt.Errorf("附件大小无效")
		}
		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("附件编号重复")
		}
		seen[item.ID] = struct{}{}
	}
	return nil
}

func cloneAttachments(items []Attachment) []Attachment {
	if items == nil {
		return []Attachment{}
	}
	return append([]Attachment(nil), items...)
}

func cloneEquipment(item Equipment) Equipment {
	item.Attachments = cloneAttachments(item.Attachments)
	return item
}

func cloneLogs(items []OperationLog) []OperationLog {
	return append([]OperationLog(nil), items...)
}
