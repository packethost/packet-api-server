package store

import (
	"github.com/google/uuid"
	"github.com/packethost/packet-api-server/pkg/util"
	"github.com/packethost/packngo"
)

const (
	// we use constants for these, since in memory, it does not really matter
	iqn = "iqn.2013-05.com.daterainc:tc:01:sn:73d3e29022fddba4"
	ip1 = "10.144.32.8"
	ip2 = "10.144.48.8"
)

// Memory is an implementation of DataStore which stores everything in memory
type Memory struct {
	volumes     map[string]*packngo.Volume
	attachments map[string]*packngo.VolumeAttachment
	facilities  []*packngo.Facility
	devices     map[string]*packngo.Device
}

// ListFacilities returns facilities; if blank, it knows about ewr1
func (m *Memory) ListFacilities() ([]*packngo.Facility, error) {
	if len(m.facilities) != 0 {
		return m.facilities, nil
	}
	return []*packngo.Facility{
		{ID: "e1e9c52e-a0bc-4117-b996-0fc94843ea09", Name: "Parsippany, NJ", Code: "ewr1"},
	}, nil

}

// CreateDevice creates a new device
func (m *Memory) CreateDevice(projectID, name string) (*packngo.Device, error) {
	device := &packngo.Device{
		DeviceRaw: packngo.DeviceRaw{
			ID:       uuid.New().String(),
			Hostname: name,
			State:    "active",
		},
	}
	m.devices[device.ID] = device
	return device, nil
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

// ListVolumes list the volumes for the project
func (m *Memory) ListVolumes(projectID string, listOpt *packngo.ListOptions) ([]*packngo.Volume, error) {
	count := len(m.volumes)
	vols := make([]*packngo.Volume, 0, count)
	for _, v := range m.volumes {
		if len(vols) >= count {
			break
		}
		vols = append(vols, v)
	}
	return vols, nil
}

// GetVolume get information about a single volume
func (m *Memory) GetVolume(volID string) (*packngo.Volume, error) {
	if vol, ok := m.volumes[volID]; ok {
		return vol, nil
	}
	return nil, nil
}

// DeleteVolume delete a single volume
func (m *Memory) DeleteVolume(volID string) (bool, error) {
	if _, ok := m.volumes[volID]; ok {
		delete(m.volumes, volID)
		return true, nil
	}
	return false, nil
}

// CreateVolume create a new volume
func (m *Memory) CreateVolume(cvr packngo.VolumeCreateRequest) (*packngo.Volume, error) {
	// just create it
	uuid := uuid.New().String()
	vol := &packngo.Volume{
		ID:          uuid,
		Name:        util.VolumeIDToName(uuid),
		Description: cvr.Description,
		Size:        cvr.Size,
		State:       "active",
		Plan:        &packngo.Plan{ID: cvr.PlanID},
	}
	m.volumes[uuid] = vol
	return vol, nil
}

// AttachVolume attach a volume to a device
func (m *Memory) AttachVolume(volID string, deviceID string) (*packngo.VolumeAttachment, error) {
	var (
		vol *packngo.Volume
		dev *packngo.Device
		ok  bool
	)
	// make sure we have the volume and dveice
	if vol, ok = m.volumes[volID]; !ok {
		return nil, nil
	}
	if dev, ok = m.devices[deviceID]; !ok {
		return nil, nil
	}
	uuid := uuid.New().String()
	attachment := packngo.VolumeAttachment{
		ID:     uuid,
		Device: packngo.Device{DeviceRaw: packngo.DeviceRaw{ID: deviceID}},
		Volume: *vol,
	}
	vol.Attachments = append(vol.Attachments, &attachment)
	dev.Volumes = append(dev.Volumes, vol)
	return &attachment, nil
}

// DetachVolume detach a volume from a device
func (m *Memory) DetachVolume(attachID string) (bool, error) {
	var (
		attachment *packngo.VolumeAttachment
		ok         bool
	)
	if attachment, ok = m.attachments[attachID]; !ok {
		return false, nil
	}
	devID := attachment.Device.ID
	volID := attachment.Volume.ID
	// remove the attachment from the volume
	if vol, ok := m.volumes[volID]; ok {
		n := 0
		for _, x := range vol.Attachments {
			if x.ID != attachID {
				vol.Attachments[n] = x
				n++
			}
		}
		vol.Attachments = vol.Attachments[:n]
	}
	// remove the volume from the device
	if dev, ok := m.devices[devID]; ok {
		n := 0
		for _, x := range dev.Volumes {
			if x.ID != volID {
				dev.Volumes[n] = x
				n++
			}
		}
		dev.Volumes = dev.Volumes[:n]
	}

	// delete the attachment
	delete(m.attachments, attachID)

	return true, nil
}

// GetAttachmentMetadata get the metadata about a given attachment
func (m *Memory) GetAttachmentMetadata(attachID string) (string, []string, error) {
	return iqn, []string{ip1, ip2}, nil
}
