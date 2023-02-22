package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/packethost/packet-api-server/pkg/store"
	"github.com/packethost/packngo"
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
	// create a BGP config for a project
	r.HandleFunc("/projects/{projectID}/bgp-configs", p.createBGPHandler).Methods("POST")
	// get all devices for a project
	r.HandleFunc("/projects/{projectID}/devices", p.listDevicesHandler).Methods("GET")
	// get a single device
	r.HandleFunc("/devices/{deviceID}", p.getDeviceHandler).Methods("GET")
	// handle metadata requests
	return r
}

// list all facilities
func (p *PacketServer) allFacilitiesHandler(w http.ResponseWriter, r *http.Request) {
	facilities, err := p.Store.ListFacilities()
	if err != nil {
		p.ErrorHandler.Error(err)
	}
	var resp = struct {
		Facilities []*packngo.Facility `json:"facilities"`
	}{
		Facilities: facilities,
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

// createBGPHandler enable BGP for a project
func (p *PacketServer) createBGPHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectID"]
	// get the info from the body
	decoder := json.NewDecoder(r.Body)
	var cbgpcr packngo.CreateBGPConfigRequest
	err := decoder.Decode(&cbgpcr)
	if err != nil {
		p.ErrorHandler.Error(err)
	}

	if err := p.Store.EnableBGP(projectID, cbgpcr); err != nil {
		p.ErrorHandler.Error(err)
	}
	w.WriteHeader(http.StatusCreated)
}
