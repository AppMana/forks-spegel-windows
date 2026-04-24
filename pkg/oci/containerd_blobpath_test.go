package oci

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestBlobPathLayout asserts the on-disk blob path matches what containerd
// actually writes on both Linux and Windows. The Linux layout was already
// implicitly assumed by the existing codebase; the Windows layout was
// captured live from appmana-023 (containerd 2.2.1, Windows Server 2022) —
// see probes/appmana-023/findings.md and the fixture at
// testdata/windows-probe/appmana-023.txt. Both layouts use the same
// structure ("<content>/blobs/<algo>/<digest>"); the only divergence is the
// path separator, which filepath.Join handles natively.
func TestBlobPathLayout(t *testing.T) {
	const digest = "0002648c324f375c764bd251d4dacb49fcf77ba09f3711e15b423862eb7b0824"
	cases := []struct {
		name    string
		content string
		want    string
		onlyOn  string
	}{
		{
			name:    "linux",
			content: "/var/lib/containerd/io.containerd.content.v1.content",
			want:    "/var/lib/containerd/io.containerd.content.v1.content/blobs/sha256/" + digest,
			onlyOn:  "linux",
		},
		{
			name:    "windows",
			content: `C:\ProgramData\containerd\root\io.containerd.content.v1.content`,
			want:    `C:\ProgramData\containerd\root\io.containerd.content.v1.content\blobs\sha256\` + digest,
			onlyOn:  "windows",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.onlyOn != runtime.GOOS {
				t.Skipf("layout only valid on %s", tc.onlyOn)
			}
			got := filepath.Join(tc.content, "blobs", "sha256", digest)
			if got != tc.want {
				t.Errorf("blob path = %q, want %q", got, tc.want)
			}
		})
	}
}
