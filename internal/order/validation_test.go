package order

import "testing"

func TestClassify(t *testing.T) {
	tests := []struct {
		name       string
		statusCode *int
		status     *string
		want       ValidationClass
	}{
		{name: "migrated", statusCode: intPointer(1), status: stringPointer("accepted"), want: ValidationMigrated},
		{name: "wrong mapped value", statusCode: intPointer(1), status: stringPointer("shipped"), want: ValidationInconsistent},
		{name: "known code missing status", statusCode: intPointer(1), status: nil, want: ValidationInconsistent},
		{name: "unknown code missing status", statusCode: intPointer(9), status: nil, want: ValidationUnknown},
		{name: "null code missing status", statusCode: nil, status: nil, want: ValidationUnknown},
		{name: "unknown code has status", statusCode: intPointer(9), status: stringPointer("accepted"), want: ValidationInconsistent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.statusCode, tt.status); got != tt.want {
				t.Fatalf("Classify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
