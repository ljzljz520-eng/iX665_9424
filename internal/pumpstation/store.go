package pumpstation

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

var ErrNotFound = errors.New("设备档案不存在")

type AttachmentStorage interface {
	Save(equipmentID string, items []Attachment) error
	List(equipmentID string) []Attachment
}

type FixtureAttachmentStorage struct {
	mu          sync.RWMutex
	items       map[string][]Attachment
	failFileKey string
}

func NewFixtureAttachmentStorage() *FixtureAttachmentStorage {
	return &FixtureAttachmentStorage{items: make(map[string][]Attachment)}
}

func (s *FixtureAttachmentStorage) Save(equipmentID string, items []Attachment) error {
	for _, item := range items {
		if item.StorageKey == s.failFileKey || item.FileName == s.failFileKey {
			return errors.New("附件存储失败")
		}
	}
	s.mu.Lock()
	s.items[equipmentID] = cloneAttachments(items)
	s.mu.Unlock()
	return nil
}

func (s *FixtureAttachmentStorage) List(equipmentID string) []Attachment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneAttachments(s.items[equipmentID])
}

func (s *FixtureAttachmentStorage) FailOn(value string) {
	s.mu.Lock()
	s.failFileKey = value
	s.mu.Unlock()
}

type memoryStore struct {
	mu          sync.RWMutex
	equipment   map[string]Equipment
	logs        map[string][]OperationLog
	nextID      int64
	attachments AttachmentStorage
}

func newMemoryStore(storage AttachmentStorage) *memoryStore {
	return &memoryStore{
		equipment:   make(map[string]Equipment),
		logs:        make(map[string][]OperationLog),
		nextID:      1,
		attachments: storage,
	}
}

func (s *memoryStore) addLog(equipmentID, action, detail, result string) OperationLog {
	entry := OperationLog{ID: s.nextID, EquipmentID: equipmentID, Action: action, Detail: detail, Result: result}
	s.nextID++
	s.logs[equipmentID] = append(s.logs[equipmentID], entry)
	return entry
}

func (s *memoryStore) writeLog(equipmentID, action, detail, result string) error {
	s.addLog(equipmentID, action, detail, result)
	return nil
}

func (s *memoryStore) snapshot(equipmentID string) (Equipment, []OperationLog, bool) {
	equipment, ok := s.equipment[equipmentID]
	if !ok {
		return Equipment{}, nil, false
	}
	return cloneEquipment(equipment), cloneLogs(s.logs[equipmentID]), true
}

func (s *memoryStore) restore(equipmentID string, equipment Equipment, logs []OperationLog, nextID int64) {
	s.equipment[equipmentID] = cloneEquipment(equipment)
	s.logs[equipmentID] = cloneLogs(logs)
	s.nextID = nextID
}

func (s *memoryStore) list(query string, page, pageSize int) Page {
	items := make([]Equipment, 0, len(s.equipment))
	needle := strings.ToLower(strings.TrimSpace(query))
	for _, item := range s.equipment {
		if needle == "" || strings.Contains(strings.ToLower(item.StationName), needle) || strings.Contains(strings.ToLower(item.EquipmentNumber), needle) || strings.Contains(strings.ToLower(item.ResponsibleTeam), needle) {
			items = append(items, cloneEquipment(item))
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].EquipmentNumber < items[j].EquipmentNumber })
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	total := len(items)
	totalPages := (total + pageSize - 1) / pageSize
	start := (page - 1) * pageSize
	if start >= total {
		return Page{Items: []Equipment{}, Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return Page{Items: items[start:end], Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages}
}
