package file

import "testing"

var (
	mockedExistsFunc = func(path string) (bool, error) {
		if path == "photo.jpg" {
			return true, nil
		}

		if path == "photo_1.jpg" {
			return true, nil
		}

		if path == "image.jpg" {
			return true, nil
		}

		return false, nil
	}

	mockedSumFunc = func(path string) (string, error) {
		return "some_checksum", nil
	}
)

func TestHandler_IsDuplicated_NewFile(t *testing.T) {
	duplicated, newPath, _ := NewHandler(mockedExistsFunc, mockedSumFunc).
		IsDuplicated("image.png", "some_checksum")

	if duplicated {
		t.Error("invalid duplicated value")
	}

	if newPath != "image.png" {
		t.Error("invalid new path value")
	}
}

func TestHandler_IsDuplicated_Exists(t *testing.T) {
	duplicated, newPath, _ := NewHandler(mockedExistsFunc, mockedSumFunc).
		IsDuplicated("photo.jpg", "some_checksum")

	if !duplicated {
		t.Error("invalid duplicated value")
	}

	if newPath != "" {
		t.Error("invalid new path value")
	}
}

func TestHandler_IsDuplicated_SameName_1(t *testing.T) {
	duplicated, newPath, _ := NewHandler(mockedExistsFunc, mockedSumFunc).
		IsDuplicated("image.jpg", "another_checksum")

	if duplicated {
		t.Error("invalid duplicated value")
	}

	if newPath != "image_1.jpg" {
		t.Error("invalid new path value")
	}
}

func TestHandler_IsDuplicated_SameName_2(t *testing.T) {
	duplicated, newPath, _ := NewHandler(mockedExistsFunc, mockedSumFunc).
		IsDuplicated("photo.jpg", "another_checksum")

	if duplicated {
		t.Error("invalid duplicated value")
	}

	if newPath != "photo_2.jpg" {
		t.Error("invalid new path value")
	}
}
