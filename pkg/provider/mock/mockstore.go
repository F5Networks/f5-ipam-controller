package mock

type MockDBStore struct {
	Data MockData
}

type MockData struct {
	IPAMLabelMap map[string]string
	CleanUpFlag  bool
	Hostdata     map[string]string
	IpList       []string
	index        int
	LabelData    map[string]string
}

func MockNewStore(data MockData) *MockDBStore {
	return &MockDBStore{Data: data}
}

func (mockStore *MockDBStore) CreateTables() bool {
	return true
}

func (mockStore *MockDBStore) InsertIP(ips []string, ipamLabel string) {
}

func (mockStore *MockDBStore) DisplayIPRecords() {
}

func (mockStore *MockDBStore) AllocateIP(ipamLabel string) string {
	ip := mockStore.Data.IpList[mockStore.Data.index]
	mockStore.Data.LabelData[ipamLabel] = ip
	mockStore.Data.index++
	return ip
}

func (mockStore *MockDBStore) GetIPAddress(ipamLabel, hostname string) string {
	return mockStore.Data.LabelData[ipamLabel]
}

func (mockStore *MockDBStore) ReleaseIP(ip string) {
	for k, v := range mockStore.Data.LabelData {
		if v == ip {
			delete(mockStore.Data.LabelData, k)
		}
	}
	mockStore.Data.index--
}

func (mockStore *MockDBStore) CreateARecord(hostname, ipAddr string) bool {
	mockStore.Data.Hostdata[hostname] = ipAddr
	return true
}

func (mockStore *MockDBStore) DeleteARecord(hostname, ipAddr string) bool {
	delete(mockStore.Data.Hostdata, hostname)
	return true
}

func (mockStore *MockDBStore) GetLabelMap() map[string]string {
	return mockStore.Data.IPAMLabelMap
}

func (mockStore *MockDBStore) AddLabel(label, ipRange string) bool {
	return true
}

func (mockStore *MockDBStore) RemoveLabel(label string) bool {
	return true
}

func (mockStore *MockDBStore) CleanUpLabel(label string) {
	mockStore.Data.CleanUpFlag = true
}
