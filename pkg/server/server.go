package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/packethost/packet-api-server/pkg/store"
	"github.com/packethost/packet-api-server/pkg/util"
	"github.com/packethost/packngo"
	"github.com/packethost/packngo/metadata"
)

// ErrorHandler a handler for errors that can choose to exit or not
// if it wants, it can exit entirely
type ErrorHandler interface {
	Error(error)
}

// PacketServer a handler creator for an http server
type PacketServer struct {
	Store store.DataStore
	ErrorHandler
	// MetadataDevice ID of the device whose metadata we will serve
	MetadataDevice string
}

// CreateHandler create an http.Handler
func (p *PacketServer) CreateHandler() http.Handler {
	r := mux.NewRouter()
	// list all facilities
	r.HandleFunc("/facilities", p.allFacilitiesHandler).Methods("GET")
	// get all volumes for a project
	r.HandleFunc("/projects/{projectID}/storage", p.listVolumesHandler).Methods("GET")
	// get information about a specific volume
	r.HandleFunc("/storage/{volumeID}", p.getVolumeHandler).Methods("GET")
	// create a volume for a project
	r.HandleFunc("/projects/{projectID}/storage", p.createVolumeHandler).Methods("POST")
	// delete a volume
	r.HandleFunc("/storage/{volumeID}", p.deleteVolumeHandler).Methods("DELETE")
	// attach a volume to a host
	r.HandleFunc("/storage/{volumeID}/attachments", p.volumeAttachHandler).Methods("POST")
	// detach a volume from a host
	r.HandleFunc("/storage/attachments/{attachmentID}", p.volumeDetachHandler).Methods("DELETE")
	// get all devices for a project
	r.HandleFunc("/projects/{projectID}/devices", p.listDevicesHandler).Methods("GET")
	// get a single device
	r.HandleFunc("/devices/{deviceID}", p.getDeviceHandler).Methods("GET")
	// handle metadata requests
	r.HandleFunc("/metadata", p.metadataHandler).Methods("GET")
	return r
}

// list all facilities
func (p *PacketServer) allFacilitiesHandler(w http.ResponseWriter, r *http.Request) {
	facilities, err := p.Store.ListFacilities()
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	var resp = struct {
		facilities []*packngo.Facility
	}{
		facilities: facilities,
	}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
}

// list all devices for a project
func (p *PacketServer) listDevicesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectID"]

	devices, err := p.Store.ListDevices(projectID)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	var resp = struct {
		Devices []*packngo.Device `json:"devices"`
	}{
		Devices: devices,
	}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
}

// get information about a specific device
func (p *PacketServer) getDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volID := vars["deviceID"]
	dev, err := p.Store.GetDevice(volID)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	if dev != nil {
		err := json.NewEncoder(w).Encode(&dev)
		if err != nil {
			p.ErrorHandler.Error(err)
		}
		return
	}
	http.NotFound(w, r)
}

// list all volumes for a project
func (p *PacketServer) listVolumesHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := mux.Vars(r)
	projectID := vars["projectID"]

	// were we asked to limit it?
	listOpt, err := paramsToListOpts(r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusOK)
		resp := struct {
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("invalid query parameters: %v", err),
		}
		err = json.NewEncoder(w).Encode(&resp)
		if err != nil {
			p.ErrorHandler.Error(err)
		}
		return
	}
	vols, err := p.Store.ListVolumes(projectID, listOpt)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	var resp = struct {
		Volumes []*packngo.Volume `json:"volumes"`
	}{
		Volumes: vols,
	}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
}

// get information about a specific volume
func (p *PacketServer) getVolumeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volID := vars["volumeID"]
	vol, err := p.Store.GetVolume(volID)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	if vol != nil {
		err := json.NewEncoder(w).Encode(&vol)
		if err != nil {
			p.ErrorHandler.Error(err)
		}
		return
	}
	http.NotFound(w, r)
}

// delete a volume
func (p *PacketServer) deleteVolumeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volID := vars["volumeID"]
	deleted, err := p.Store.DeleteVolume(volID)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	if deleted {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.NotFound(w, r)
}

// create a volume
func (p *PacketServer) createVolumeHandler(w http.ResponseWriter, r *http.Request) {
	// get the info from the body
	decoder := json.NewDecoder(r.Body)
	var cvr packngo.VolumeCreateRequest
	err := decoder.Decode(&cvr)
	if err != nil {
		p.ErrorHandler.Error(err)
	}

	vol, err := p.Store.CreateVolume(cvr)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&vol)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
}

// attach volume
func (p *PacketServer) volumeAttachHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volID := vars["volumeID"]
	// get the device from the body
	decoder := json.NewDecoder(r.Body)
	var attachRequest struct {
		Device string `json:"device_id"`
	}
	err := decoder.Decode(&attachRequest)
	if err != nil {
		p.ErrorHandler.Error(err)
	}

	attachment, err := p.Store.AttachVolume(volID, attachRequest.Device)
	if err != nil {
		p.ErrorHandler.Error(err)
	}

	if attachment != nil {
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(&attachment)
		if err != nil {
			p.ErrorHandler.Error(err)
		}
		return
	}
	http.NotFound(w, r)
}

// detach volume
func (p *PacketServer) volumeDetachHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attachID := vars["attachmentID"]
	detached, err := p.Store.DetachVolume(attachID)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	if detached {
		w.WriteHeader(http.StatusOK)
	}
	http.NotFound(w, r)
}

// metadata handler
// this is a difficult one, since we are building a generic API server
// but the metadata service is meant to assume it already knows the device ID
// and it cannot request it from the client sending the http request

// for now, we pre-define it for a device to start
func (p *PacketServer) metadataHandler(w http.ResponseWriter, r *http.Request) {
	// do not do anything if we do not have a node defined
	if p.MetadataDevice == "" {
		http.NotFound(w, r)
		return
	}
	// now find the given device
	dev, err := p.Store.GetDevice(p.MetadataDevice)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	// find the volumes for the given device, find the attachments, get the metadata
	allVols := dev.Volumes
	vols := make([]*metadata.VolumeInfo, 0, len(allVols))
	for _, v := range allVols {
		attachment := v.Attachments[0]
		iqn, ips, err := p.Store.GetAttachmentMetadata(attachment.ID)
		if err != nil {
			p.ErrorHandler.Error(err)
		}
		vols = append(vols, &metadata.VolumeInfo{
			Name: util.VolumeIDToName(v.ID),
			IQN:  iqn,
			IPs:  []net.IP{net.ParseIP(ips[0]), net.ParseIP(ips[1])},
			Capacity: struct {
				Size int    `json:"size,string"`
				Unit string `json:"unit"`
			}{
				Size: v.Size,
				Unit: "gb",
			},
		})
	}
	var resp = struct {
		Volumes []*metadata.VolumeInfo `json:"volumes"`
	}{
		Volumes: vols,
	}
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		p.ErrorHandler.Error(err)
	}
}

func paramsToListOpts(params url.Values) (*packngo.ListOptions, error) {
	listOpt := &packngo.ListOptions{}
	perPage, ok := params["per_page"]
	if ok && len(perPage) > 0 && perPage[0] != "" {
		count, err := strconv.Atoi(perPage[0])
		// any error converting should be returned
		if err != nil {
			return nil, fmt.Errorf("error converting per_page %s to int: %v", perPage[0], err)
		}
		listOpt.PerPage = count
	}
	page, ok := params["page"]
	if ok && len(page) > 0 && page[0] != "" {
		pageNo, err := strconv.Atoi(page[0])
		// any error converting should be returned
		if err != nil {
			return nil, fmt.Errorf("error converting page %s to int: %v", page[0], err)
		}
		listOpt.Page = pageNo
	}
	return listOpt, nil
}
