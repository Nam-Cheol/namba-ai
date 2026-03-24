package namba

import "testing"

func TestReleaseAssetName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{
			name:   "windows 386",
			goos:   "windows",
			goarch: "386",
			want:   "namba_Windows_x86.zip",
		},
		{
			name:   "windows amd64",
			goos:   "windows",
			goarch: "amd64",
			want:   "namba_Windows_x86_64.zip",
		},
		{
			name:   "windows arm64",
			goos:   "windows",
			goarch: "arm64",
			want:   "namba_Windows_arm64.zip",
		},
		{
			name:   "linux amd64",
			goos:   "linux",
			goarch: "amd64",
			want:   "namba_Linux_x86_64.tar.gz",
		},
		{
			name:   "mac arm64",
			goos:   "darwin",
			goarch: "arm64",
			want:   "namba_macOS_arm64.tar.gz",
		},
		{
			name:    "unsupported",
			goos:    "freebsd",
			goarch:  "amd64",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := releaseAssetName(tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s/%s", tt.goos, tt.goarch)
				}
				return
			}
			if err != nil {
				t.Fatalf("releaseAssetName(%q, %q) returned error: %v", tt.goos, tt.goarch, err)
			}
			if got != tt.want {
				t.Fatalf("releaseAssetName(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}
