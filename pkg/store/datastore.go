package store

import (
	"github.com/packethost/packngo"
)

// DataStore is the item that retrieves backend information to serve out
// following a contract API
type DataStore interface {
	CreateFacility(name, code string) (*packngo.Facility, error)
	ListFacilities() ([]*packngo.Facility, error)
	GetFacility(id string) (*packngo.Facility, error)
	GetFacilityByCode(code string) (*packngo.Facility, error)
	CreatePlan(slug, name string) (*packngo.Plan, error)
	GetPlan(planID string) (*packngo.Plan, error)
	GetPlanBySlug(slug string) (*packngo.Plan, error)
	CreateDevice(projectID, name string, plan *packngo.Plan, facility *packngo.Facility) (*packngo.Device, error)
	ListDevices(projectID string) ([]*packngo.Device, error)
	GetDevice(deviceID string) (*packngo.Device, error)
	DeleteDevice(deviceID string) (bool, error)
	ListVolumes(projectID string, listOpt *packngo.ListOptions) ([]*packngo.Volume, error)
	GetVolume(volID string) (*packngo.Volume, error)
	DeleteVolume(volID string) (bool, error)
	CreateVolume(cvr packngo.VolumeCreateRequest) (*packngo.Volume, error)
	AttachVolume(volID string, deviceID string) (*packngo.VolumeAttachment, error)
	DetachVolume(attachID string) (bool, error)
	GetAttachmentMetadata(attachID string) (string, []string, error)
}
