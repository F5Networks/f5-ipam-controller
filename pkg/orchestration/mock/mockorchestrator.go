package mock

import "github.com/F5Networks/f5-ipam-controller/pkg/ipamspec"

var SetupCalled, StartCalled, StopCalled bool

type MockOrch struct {
	// Channel for sending request to controller
	ReqChan chan<- ipamspec.IPAMRequest
	// Channel for receiving responce from controller
	RespChan <-chan ipamspec.IPAMResponse
}

func (moc *MockOrch) SetupCommunicationChannels(reqChan chan<- ipamspec.IPAMRequest, respChan <-chan ipamspec.IPAMResponse) {
	SetupCalled = true
	{
		moc.ReqChan = reqChan
		moc.RespChan = respChan
	}
}
func (moc *MockOrch) Stop() {
	StopCalled = true
}
func (moc *MockOrch) Start(stopCh <-chan struct{}) {
	StartCalled = true
}
