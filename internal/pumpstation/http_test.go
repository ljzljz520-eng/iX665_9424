package pumpstation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func requestJSON(t *testing.T, handler http.Handler, method, path string, value any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("request body error = %v", err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestEquipmentHTTPWorkflow(t *testing.T) {
	handler := NewHTTPHandler(NewFixtureService())
	search := httptest.NewRecorder()
	handler.ServeHTTP(search, httptest.NewRequest(http.MethodGet, "/api/equipment?query=西郊&page=1&page_size=10", nil))
	if search.Code != http.StatusOK {
		t.Fatalf("search status = %d", search.Code)
	}
	var page Page
	if err := json.NewDecoder(search.Body).Decode(&page); err != nil {
		t.Fatalf("search response error = %v", err)
	}
	if page.Total != 1 || page.Items[0].EquipmentNumber != "P-2001" {
		t.Errorf("search page = %#v", page)
	}

	created := requestJSON(t, handler, http.MethodPost, "/api/equipment", CreateRequest{
		StationName:      "北区清水泵站",
		EquipmentNumber:  "P-4001",
		Power:            45,
		InstallationSite: "泵房C区",
		ResponsibleTeam:  "运行四班",
		CommissionedOn:   "2024-02-01",
		RunningStatus:    StatusRunning,
		Attachments:      []Attachment{},
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d", created.Code)
	}
	var item Equipment
	if err := json.NewDecoder(created.Body).Decode(&item); err != nil {
		t.Fatalf("create response error = %v", err)
	}
	edited := requestJSON(t, handler, http.MethodPut, "/api/equipment/"+item.ID, EditRequest{
		StationName:      item.StationName,
		EquipmentNumber:  item.EquipmentNumber,
		Power:            48,
		InstallationSite: item.InstallationSite,
		ResponsibleTeam:  item.ResponsibleTeam,
		CommissionedOn:   item.CommissionedOn,
		RunningStatus:    StatusStopped,
		Attachments:      []Attachment{},
	})
	if edited.Code != http.StatusOK {
		t.Fatalf("edit status = %d", edited.Code)
	}
	attachmentUpdate := requestJSON(t, handler, http.MethodPut, "/api/equipment/"+item.ID+"/attachments", attachmentUpdateRequest{Attachments: []Attachment{fixtureAttachment("7d444840-9dc0-11d1-b245-5ffdce74fad5", "巡检记录.pdf", "fixture/p-4001/inspection.pdf")}})
	if attachmentUpdate.Code != http.StatusOK {
		t.Fatalf("attachment update status = %d", attachmentUpdate.Code)
	}

	detail := httptest.NewRecorder()
	handler.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/equipment/"+item.ID, nil))
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d", detail.Code)
	}
	var updated Equipment
	if err := json.NewDecoder(detail.Body).Decode(&updated); err != nil {
		t.Fatalf("detail response error = %v", err)
	}
	if updated.Power != 48 || updated.RunningStatus != StatusStopped || len(updated.Attachments) != 1 {
		t.Errorf("detail = %#v", updated)
	}

	logs := httptest.NewRecorder()
	handler.ServeHTTP(logs, httptest.NewRequest(http.MethodGet, "/api/equipment/"+item.ID+"/logs", nil))
	if logs.Code != http.StatusOK {
		t.Fatalf("logs status = %d", logs.Code)
	}
	var logPage struct {
		Items []OperationLog `json:"items"`
	}
	if err := json.NewDecoder(logs.Body).Decode(&logPage); err != nil {
		t.Fatalf("logs response error = %v", err)
	}
	if len(logPage.Items) != 3 {
		t.Errorf("logs = %#v", logPage.Items)
	}
}

func TestEquipmentHTTPRejectsInvalidStatus(t *testing.T) {
	handler := NewHTTPHandler(NewFixtureService())
	response := requestJSON(t, handler, http.MethodPut, "/api/equipment/EQ-001", EditRequest{
		StationName:      "东城净水厂一号泵站",
		EquipmentNumber:  "P-1001",
		Power:            75,
		InstallationSite: "一号机房A列",
		ResponsibleTeam:  "运行一班",
		CommissionedOn:   "2022-04-18",
		RunningStatus:    "未知",
		Attachments:      []Attachment{},
	})
	if response.Code != http.StatusBadRequest {
		t.Errorf("invalid status response = %d", response.Code)
	}
}
