package pumpstation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type HTTPHandler struct {
	service *Service
}

type attachmentUpdateRequest struct {
	Attachments []Attachment `json:"attachments"`
}

func NewHTTPHandler(service *Service) http.Handler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/healthz" {
		writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
		return
	}
	if request.URL.Path == "/api/equipment" {
		h.collection(writer, request)
		return
	}
	prefix := "/api/equipment/"
	if !strings.HasPrefix(request.URL.Path, prefix) {
		writeError(writer, http.StatusNotFound, "接口不存在")
		return
	}
	path := strings.TrimPrefix(request.URL.Path, prefix)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(writer, http.StatusNotFound, "设备档案不存在")
		return
	}
	id := parts[0]
	switch {
	case len(parts) == 1:
		h.item(writer, request, id)
	case len(parts) == 2 && parts[1] == "attachments":
		h.attachments(writer, request, id)
	case len(parts) == 2 && parts[1] == "logs" && request.Method == http.MethodGet:
		h.logs(writer, id)
	default:
		writeError(writer, http.StatusNotFound, "接口不存在")
	}
}

func (h *HTTPHandler) collection(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		page := queryInt(request, "page", 1)
		pageSize := queryInt(request, "page_size", 10)
		writeJSON(writer, http.StatusOK, h.service.Search(request.URL.Query().Get("query"), page, pageSize))
	case http.MethodPost:
		var payload CreateRequest
		if !decodeJSON(writer, request, &payload) {
			return
		}
		item, err := h.service.Create(payload)
		if err != nil {
			writeServiceError(writer, err)
			return
		}
		writeJSON(writer, http.StatusCreated, item)
	default:
		writeError(writer, http.StatusMethodNotAllowed, "请求方法不支持")
	}
}

func (h *HTTPHandler) item(writer http.ResponseWriter, request *http.Request, id string) {
	switch request.Method {
	case http.MethodGet:
		item, err := h.service.Detail(id)
		if err != nil {
			writeServiceError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, item)
	case http.MethodPut:
		var payload EditRequest
		if !decodeJSON(writer, request, &payload) {
			return
		}
		item, err := h.service.Update(id, payload)
		if err != nil {
			writeServiceError(writer, err)
			return
		}
		writeJSON(writer, http.StatusOK, item)
	default:
		writeError(writer, http.StatusMethodNotAllowed, "请求方法不支持")
	}
}

func (h *HTTPHandler) attachments(writer http.ResponseWriter, request *http.Request, id string) {
	if request.Method != http.MethodPut {
		writeError(writer, http.StatusMethodNotAllowed, "请求方法不支持")
		return
	}
	var payload attachmentUpdateRequest
	if !decodeJSON(writer, request, &payload) {
		return
	}
	item, err := h.service.UpdateAttachments(id, payload.Attachments)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, item)
}

func (h *HTTPHandler) logs(writer http.ResponseWriter, id string) {
	items, err := h.service.Logs(id)
	if err != nil {
		writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"items": items})
}

func queryInt(request *http.Request, key string, fallback int) int {
	value, err := strconv.Atoi(request.URL.Query().Get(key))
	if err != nil {
		return fallback
	}
	return value
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, target any) bool {
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(target); err != nil {
		writeError(writer, http.StatusBadRequest, "请求数据格式无效")
		return false
	}
	return true
}

func writeServiceError(writer http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, ErrNotFound) {
		status = http.StatusNotFound
	}
	writeError(writer, status, err.Error())
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"error": message})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
