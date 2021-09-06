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

func NewMockStore(data MockData) *MockDBStore {
	return &MockDBStore{Data: data}
}

func (ms *MockDBStore) CreateTables() bool {
	return true
}

func (ms *MockDBStore) InsertIPs(ips []string, ipamLabel string) {
}

func (ms *MockDBStore) DisplayIPRecords() {
}

func (ms *MockDBStore) AllocateIP(ipamLabel, reference string) string {
	ip := ms.Data.IpList[ms.Data.index]
	ms.Data.LabelData[ipamLabel] = ip
	ms.Data.index++
	return ip
}

func (ms *MockDBStore) ReleaseIP(ip string) {
	for k, v := range ms.Data.LabelData {
		if v == ip {
			delete(ms.Data.LabelData, k)
		}
	}
	ms.Data.index--
}

func (ms *MockDBStore) GetIPAddressFromARecord(ipamLabel, hostname string) string {
	return ms.Data.LabelData[ipamLabel]
}

func (ms *MockDBStore) GetIPAddressFromReference(ipamLabel, reference string) string {
	return ms.Data.LabelData[ipamLabel]
}

func (ms *MockDBStore) CreateARecord(hostname, ipAddr string) bool {
	ms.Data.Hostdata[hostname] = ipAddr
	return true
}

func (ms *MockDBStore) DeleteARecord(hostname, ipAddr string) bool {
	delete(ms.Data.Hostdata, hostname)
	return true
}

func (ms *MockDBStore) GetLabelMap() map[string]string {
	return ms.Data.IPAMLabelMap
}

func (ms *MockDBStore) AddLabel(label, ipRange string) bool {
	return true
}

func (ms *MockDBStore) RemoveLabel(label string) bool {
	return true
}

func (ms *MockDBStore) CleanUpLabel(label string) {
	ms.Data.CleanUpFlag = true
}
