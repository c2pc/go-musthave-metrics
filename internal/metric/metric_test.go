package metric

import "testing"

func TestKey_String(t *testing.T) {
	tests := []struct {
		name string
		k    Key
		want string
	}{
		{
			name: "empty",
			k:    "",
			want: "",
		},
		{
			name: "one",
			k:    "one",
			want: "one",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.k.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
