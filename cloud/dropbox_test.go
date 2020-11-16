package cloud

import "testing"

func TestDropbox_Location(t *testing.T) {
	provider := dropbox{}
	provider.infoDataReader = func() ([]byte, error) {
		json := `{"personal": {"path": "/dropbox/path"}}`
		return []byte(json), nil
	}

	location, _ := provider.Location()
	if location != "/dropbox/path" {
		t.Error("error on retrieve provider location")
	}
}
