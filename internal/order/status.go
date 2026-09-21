package order

import "fmt"

type ReadMode string

const (
	ReadModeFallback ReadMode = "fallback"
	ReadModeNew      ReadMode = "new"
)

type WriteMode string

const (
	WriteModeLegacy WriteMode = "legacy"
	WriteModeDual   WriteMode = "dual"
)

var knownStatuses = map[string]Status{
	"accepted": {LegacyCode: 1, Code: "accepted"},
	"shipped":  {LegacyCode: 2, Code: "shipped"},
	"canceled": {LegacyCode: 3, Code: "canceled"},
}

func ParseStatus(code string) (Status, error) {
	status, ok := knownStatuses[code]
	if !ok {
		return Status{}, fmt.Errorf("unknown status: %q", code)
	}
	return status, nil
}

func StatusFromLegacy(code *int) (Status, bool) {
	if code == nil {
		return Status{}, false
	}
	switch *code {
	case 1:
		return knownStatuses["accepted"], true
	case 2:
		return knownStatuses["shipped"], true
	case 3:
		return knownStatuses["canceled"], true
	default:
		return Status{}, false
	}
}

func ParseReadMode(value string) (ReadMode, error) {
	switch ReadMode(value) {
	case ReadModeFallback:
		return ReadModeFallback, nil
	case ReadModeNew:
		return ReadModeNew, nil
	default:
		return "", fmt.Errorf("unknown READ_MODE: %q", value)
	}
}

func ParseWriteMode(value string) (WriteMode, error) {
	switch WriteMode(value) {
	case WriteModeLegacy:
		return WriteModeLegacy, nil
	case WriteModeDual:
		return WriteModeDual, nil
	default:
		return "", fmt.Errorf("unknown WRITE_MODE: %q", value)
	}
}
