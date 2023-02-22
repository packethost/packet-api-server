package store

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/packethost/packngo"
)

// Memory is an implementation of DataStore which stores everything in memory
type Memory struct {
	facilities map[string]*packngo.Facility
	devices    map[string]*packngo.Device
	plans      map[string]*packngo.Plan
	bgp        map[string]*packngo.BGPConfig
}

// NewMemory returns a properly initialized Memory
func NewMemory() *Memory {
	return &Memory{
		facilities: map[string]*packngo.Facility{},
		devices:    map[string]*packngo.Device{},
		plans:      map[string]*packngo.Plan{},
		bgp:        map[string]*packngo.BGPConfig{},
	}
}

// CreateFacility creates a new facility
func (m *Memory) CreateFacility(name, code string) (*packngo.Facility, error) {
	facility := &packngo.Facility{
		ID:   uuid.New().String(),
		Name: name,
		Code: code,
	}
	m.facilities[facility.ID] = facility
	return facility, nil
}

// ListFacilities returns facilities; if blank, it knows about ewr1
func (m *Memory) ListFacilities() ([]*packngo.Facility, error) {
	count := len(m.facilities)
	if count != 0 {
		facilities := make([]*packngo.Facility, 0, count)
		for _, v := range m.facilities {
			if len(facilities) >= count {
				break
			}
			facilities = append(facilities, v)
		}
		return facilities, nil
	}
	return []*packngo.Facility{
		{ID: "e1e9c52e-a0bc-4117-b996-0fc94843ea09", Name: "Parsippany, NJ", Code: "ewr1"},
	}, nil

}

// GetFacility get a single facility by ID
func (m *Memory) GetFacility(id string) (*packngo.Facility, error) {
	if facility, ok := m.facilities[id]; ok {
		return facility, nil
	}
	return nil, nil
}

// GetFacilityByCode get a single facility by code
func (m *Memory) GetFacilityByCode(code string) (*packngo.Facility, error) {
	for _, f := range m.facilities {
		if f.Code == code {
			return f, nil
		}
	}
	return nil, nil
}

// CreatePlan create a single plan
func (m *Memory) CreatePlan(slug, name string) (*packngo.Plan, error) {
	plan := &packngo.Plan{
		ID:   uuid.New().String(),
		Name: name,
		Slug: slug,
	}
	m.plans[plan.ID] = plan
	return plan, nil
}

// GetPlan get plan by ID
func (m *Memory) GetPlan(id string) (*packngo.Plan, error) {
	if plan, ok := m.plans[id]; ok {
		return plan, nil
	}
	return nil, nil
}

// GetPlanBySlug get plan by slug
func (m *Memory) GetPlanBySlug(slug string) (*packngo.Plan, error) {
	for _, p := range m.plans {
		if p.Slug == slug {
			return p, nil
		}
	}
	return nil, nil
}

// CreateDevice creates a new device
func (m *Memory) CreateDevice(projectID, name string, plan *packngo.Plan, facility *packngo.Facility) (*packngo.Device, error) {
	if facility == nil {
		return nil, fmt.Errorf("must include a valid facility")
	}
	if plan == nil {
		return nil, fmt.Errorf("must include a valid plan")
	}
	device := &packngo.Device{
		ID:       uuid.New().String(),
		Hostname: name,
		State:    "active",
		Facility: facility,
		Plan:     plan,
	}
	m.devices[device.ID] = device
	return device, nil
}

// UpdateDevice updates an existing device
func (m *Memory) UpdateDevice(id string, device *packngo.Device) error {
	if device == nil {
		return fmt.Errorf("must include a valid device")
	}
	if _, ok := m.devices[device.ID]; ok {
		m.devices[device.ID] = device
		return nil
	}
	return fmt.Errorf("device not found")
}

// ListDevices list all known devices for the project
func (m *Memory) ListDevices(projectID string) ([]*packngo.Device, error) {
	count := len(m.devices)
	devices := make([]*packngo.Device, 0, count)
	for _, v := range m.devices {
		if len(devices) >= count {
			break
		}
		devices = append(devices, v)
	}
	return devices, nil
}

// GetDevice get information about a single device
func (m *Memory) GetDevice(deviceID string) (*packngo.Device, error) {
	if device, ok := m.devices[deviceID]; ok {
		return device, nil
	}
	return nil, nil
}

// DeleteDevice delete a single device
func (m *Memory) DeleteDevice(deviceID string) (bool, error) {
	if _, ok := m.devices[deviceID]; ok {
		delete(m.devices, deviceID)
		return true, nil
	}
	return false, nil
}

func (m *Memory) EnableBGP(projectID string, cbgpcr packngo.CreateBGPConfigRequest) error {
	m.bgp[projectID] = &packngo.BGPConfig{
		ID:             uuid.New().String(),
		DeploymentType: cbgpcr.DeploymentType,
		Asn:            cbgpcr.Asn,
	}
	return nil
}
