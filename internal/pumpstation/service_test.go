package pumpstation

import (
	"net/http"
	"reflect"
	"testing"
)

func fixtureAttachment(id, name, key string) Attachment {
	return Attachment{ID: id, FileName: name, StorageKey: key, Size: 128}
}

func fixtureEdit(attachments []Attachment) EditRequest {
	return EditRequest{
		StationName:      "东城净水厂一号泵站",
		EquipmentNumber:  "P-1001",
		Power:            90,
		InstallationSite: "一号机房B列",
		ResponsibleTeam:  "运行一班",
		CommissionedOn:   "2022-04-18",
		RunningStatus:    StatusRunning,
		Attachments:      attachments,
	}
}

func TestEquipmentSearchAndPagination(t *testing.T) {
	service := NewFixtureService()
	page := service.Search("泵站", 2, 2)
	if page.Total != 3 {
		t.Errorf("total equipment = %d, want 3", page.Total)
	}
	if page.TotalPages != 2 {
		t.Errorf("total pages = %d, want 2", page.TotalPages)
	}
	if len(page.Items) != 1 || page.Items[0].EquipmentNumber != "P-3001" {
		t.Errorf("page items = %#v, want the third equipment", page.Items)
	}
}

func TestEquipmentDetailAndLogs(t *testing.T) {
	service := NewFixtureService()
	item, err := service.Detail("EQ-001")
	if err != nil {
		t.Fatalf("detail error = %v", err)
	}
	if item.StationName != "东城净水厂一号泵站" || len(item.Attachments) != 1 {
		t.Errorf("detail = %#v", item)
	}
	logs, err := service.Logs("EQ-001")
	if err != nil {
		t.Fatalf("logs error = %v", err)
	}
	if len(logs) != 1 || logs[0].Action != "导入设备档案" {
		t.Errorf("logs = %#v", logs)
	}
}

func TestEquipmentCreationAndEditing(t *testing.T) {
	service := NewFixtureService()
	created, err := service.Create(CreateRequest{
		StationName:      "北区清水泵站",
		EquipmentNumber:  "P-4001",
		Power:            45,
		InstallationSite: "泵房C区",
		ResponsibleTeam:  "运行四班",
		CommissionedOn:   "2024-02-01",
		RunningStatus:    StatusRunning,
		Attachments:      []Attachment{fixtureAttachment("7d444840-9dc0-11d1-b245-5ffdce74fad3", "巡检表.pdf", "fixture/p-4001/check.pdf")},
	})
	if err != nil {
		t.Fatalf("create error = %v", err)
	}
	if created.ID != "EQ-004" {
		t.Errorf("created id = %s, want EQ-004", created.ID)
	}
	updated, err := service.Update(created.ID, EditRequest{
		StationName:      "北区清水泵站",
		EquipmentNumber:  "P-4001",
		Power:            50,
		InstallationSite: "泵房D区",
		ResponsibleTeam:  "运行四班",
		CommissionedOn:   "2024-02-01",
		RunningStatus:    StatusStopped,
		Attachments:      []Attachment{},
	})
	if err != nil {
		t.Fatalf("edit error = %v", err)
	}
	if updated.Power != 50 || updated.InstallationSite != "泵房D区" || updated.RunningStatus != StatusStopped {
		t.Errorf("updated = %#v", updated)
	}
	logs, err := service.Logs(created.ID)
	if err != nil {
		t.Fatalf("logs error = %v", err)
	}
	if len(logs) != 2 || logs[1].Result != "成功" {
		t.Errorf("logs = %#v", logs)
	}
}

func TestUnsupportedRunningStatusIsRejected(t *testing.T) {
	service := NewFixtureService()
	_, err := service.Update("EQ-001", fixtureEdit(nil))
	if err != nil {
		t.Fatalf("valid edit error = %v", err)
	}
	_, err = service.Update("EQ-001", EditRequest{
		StationName:      "东城净水厂一号泵站",
		EquipmentNumber:  "P-1001",
		Power:            75,
		InstallationSite: "一号机房A列",
		ResponsibleTeam:  "运行一班",
		CommissionedOn:   "2022-04-18",
		RunningStatus:    "未知",
		Attachments:      []Attachment{},
	})
	if err == nil {
		t.Error("unsupported status was accepted")
	}
}

func TestAttachmentStorageFailureLeavesEditUnchanged(t *testing.T) {
	service := NewFixtureService()
	storage := service.AttachmentStorage().(*FixtureAttachmentStorage)
	storage.FailOn("fixture/fail-attachment.bin")
	before, err := service.Detail("EQ-001")
	if err != nil {
		t.Fatalf("before detail error = %v", err)
	}
	beforeLogs, err := service.Logs("EQ-001")
	if err != nil {
		t.Fatalf("before logs error = %v", err)
	}
	response := requestJSON(t, NewHTTPHandler(service), http.MethodPut, "/api/equipment/EQ-001", fixtureEdit([]Attachment{fixtureAttachment("7d444840-9dc0-11d1-b245-5ffdce74fad4", "失败附件.bin", "fixture/fail-attachment.bin")}))
	if response.Code != http.StatusBadRequest {
		t.Errorf("edit response after attachment storage failure = %d", response.Code)
	}
	after, detailErr := service.Detail("EQ-001")
	if detailErr != nil {
		t.Fatalf("after detail error = %v", detailErr)
	}
	if !reflect.DeepEqual(after, before) {
		t.Errorf("equipment after failed edit = %#v, want %#v", after, before)
	}
	afterLogs, logsErr := service.Logs("EQ-001")
	if logsErr != nil {
		t.Fatalf("after logs error = %v", logsErr)
	}
	if !reflect.DeepEqual(afterLogs, beforeLogs) {
		t.Errorf("logs after failed edit = %#v, want %#v", afterLogs, beforeLogs)
	}
}
