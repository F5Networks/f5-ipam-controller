package mock

import (
	"github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"
)

type MockManager struct {
	data MockData
}
type MockData struct {
	IPList      []string
	index       int
	SkipARecord bool
}

func NewMockIPAMManager(mockData MockData) (*MockManager, error) {
	return &MockManager{data: mockData}, nil
}
func (fm *MockManager) GetIPAddress(req ipamspec.IPAMRequest) string {
	if req.IPAddr != "" {
		return req.IPAddr
	}
	if req.Key == "" {
		return ""
	}
	ip := fm.data.IPList[fm.data.index]
	fm.data.index++
	return ip
}

// Creates an A record
func (fm *MockManager) CreateARecord(req ipamspec.IPAMRequest) bool {
	if req.HostName == "" || fm.data.SkipARecord {
		return false
	}
	return true
}

// Deletes an A record and releases the IP address
func (fm *MockManager) DeleteARecord(req ipamspec.IPAMRequest) {
}

// Gets and reserves the next available IP address
func (fm *MockManager) AllocateNextIPAddress(req ipamspec.IPAMRequest) string {
	if req.HostName == "" {
		return ""
	}
	ip := fm.data.IPList[fm.data.index]
	fm.data.index++
	return ip
}

// Releases an IP address
func (fm *MockManager) ReleaseIPAddress(req ipamspec.IPAMRequest) {
	fm.data.index--
}
