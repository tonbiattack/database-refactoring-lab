package order

import "testing"

func TestStatusFromLegacy(t *testing.T) {
	tests := []struct {
		name string
		code *int
		ok   bool
		want string
	}{
		{name: "accepted", code: intPointer(1), ok: true, want: "accepted"},
		{name: "shipped", code: intPointer(2), ok: true, want: "shipped"},
		{name: "canceled", code: intPointer(3), ok: true, want: "canceled"},
		{name: "unknown", code: intPointer(9), ok: false},
		{name: "null", code: nil, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := StatusFromLegacy(tt.code)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if ok && got.Code != tt.want {
				t.Fatalf("code = %q, want %q", got.Code, tt.want)
			}
		})
	}
}

func TestParseStatusRejectsUnknownValue(t *testing.T) {
	if _, err := ParseStatus("unknown"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestParseWriteModeRejectsUnknownValue(t *testing.T) {
	if _, err := ParseWriteMode("unsafe"); err == nil {
		t.Fatal("expected an error")
	}
}

func TestParseReadModeRejectsUnknownValue(t *testing.T) {
	if _, err := ParseReadMode("unsafe"); err == nil {
		t.Fatal("expected an error")
	}
}

func intPointer(value int) *int {
	return &value
}
