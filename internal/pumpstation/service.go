package pumpstation

import (
	"errors"
	"fmt"
)

type Service struct {
	store *memoryStore
}

func NewFixtureService() *Service {
	storage := NewFixtureAttachmentStorage()
	store := newMemoryStore(storage)
	seedFixture(store, storage)
	return &Service{store: store}
}

func NewService(storage AttachmentStorage) *Service {
	return &Service{store: newMemoryStore(storage)}
}

func (s *Service) AttachmentStorage() AttachmentStorage {
	return s.store.attachments
}

func (s *Service) Create(request CreateRequest) (Equipment, error) {
	id := fmt.Sprintf("EQ-%03d", len(s.store.equipment)+1)
	equipment := request.equipment(id)
	if err := validateEquipment(equipment); err != nil {
		return Equipment{}, err
	}
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	for _, existing := range s.store.equipment {
		if existing.EquipmentNumber == equipment.EquipmentNumber {
			return Equipment{}, errors.New("设备编号已存在")
		}
	}
	if err := s.store.attachments.Save(id, equipment.Attachments); err != nil {
		return Equipment{}, err
	}
	s.store.equipment[id] = cloneEquipment(equipment)
	s.store.addLog(id, "新建设备档案", "设备基础信息和附件已保存", "成功")
	return cloneEquipment(equipment), nil
}

func (s *Service) Search(query string, page, pageSize int) Page {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	return s.store.list(query, page, pageSize)
}

func (s *Service) Detail(id string) (Equipment, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	equipment, ok := s.store.equipment[id]
	if !ok {
		return Equipment{}, ErrNotFound
	}
	equipment.Attachments = s.store.attachments.List(id)
	return cloneEquipment(equipment), nil
}

func (s *Service) Logs(id string) ([]OperationLog, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()
	if _, ok := s.store.equipment[id]; !ok {
		return nil, ErrNotFound
	}
	return cloneLogs(s.store.logs[id]), nil
}

func (s *Service) Update(id string, request EditRequest) (Equipment, error) {
	updated := request.equipment(id)
	if err := validateEquipment(updated); err != nil {
		return Equipment{}, err
	}
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	previous, previousLogs, ok := s.store.snapshot(id)
	if !ok {
		return Equipment{}, ErrNotFound
	}
	previousNextID := s.store.nextID
	s.store.equipment[id] = cloneEquipment(updated)
	var err error
	err = s.store.attachments.Save(id, updated.Attachments)
	err = s.store.writeLog(id, "编辑设备档案", "附件已经更新", "成功")
	if err != nil {
		s.store.restore(id, previous, previousLogs, previousNextID)
		return Equipment{}, err
	}
	return cloneEquipment(updated), nil
}

func (s *Service) UpdateAttachments(id string, items []Attachment) (Equipment, error) {
	current, err := s.Detail(id)
	if err != nil {
		return Equipment{}, err
	}
	request := EditRequest{
		StationName:      current.StationName,
		EquipmentNumber:  current.EquipmentNumber,
		Power:            current.Power,
		InstallationSite: current.InstallationSite,
		ResponsibleTeam:  current.ResponsibleTeam,
		CommissionedOn:   current.CommissionedOn,
		RunningStatus:    current.RunningStatus,
		Attachments:      items,
	}
	return s.Update(id, request)
}

func seedFixture(store *memoryStore, storage *FixtureAttachmentStorage) {
	fixtures := []Equipment{
		{ID: "EQ-001", StationName: "东城净水厂一号泵站", EquipmentNumber: "P-1001", Power: 75, InstallationSite: "一号机房A列", ResponsibleTeam: "运行一班", CommissionedOn: "2022-04-18", RunningStatus: StatusRunning, Attachments: []Attachment{{ID: "7d444840-9dc0-11d1-b245-5ffdce74fad2", FileName: "设备铭牌.pdf", StorageKey: "fixture/p-1001/nameplate.pdf", Size: 2048}}},
		{ID: "EQ-002", StationName: "西郊原水泵站", EquipmentNumber: "P-2001", Power: 110, InstallationSite: "泵房二层", ResponsibleTeam: "维修二班", CommissionedOn: "2021-09-03", RunningStatus: StatusRepair, Attachments: []Attachment{}},
		{ID: "EQ-003", StationName: "南区加压泵站", EquipmentNumber: "P-3001", Power: 55, InstallationSite: "加压间B区", ResponsibleTeam: "运行三班", CommissionedOn: "2023-01-12", RunningStatus: StatusStopped, Attachments: []Attachment{}},
	}
	for _, item := range fixtures {
		store.equipment[item.ID] = cloneEquipment(item)
		_ = storage.Save(item.ID, item.Attachments)
		store.addLog(item.ID, "导入设备档案", "固定夹具已载入", "成功")
	}
}
