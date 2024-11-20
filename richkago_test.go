package richkago

import (
	"testing"
)

// TestDownload test main download
func TestDownload(t *testing.T) {
	controller := NewController()

	_, _, err := Download("https://mirrors.tuna.tsinghua.edu.cn/github-release/git-for-windows/git/LatestRelease/Git-2.47.0.2-64-bit.exe", "Git-2.47.0.2-64-bit.exe", controller)
	if err != nil {
		t.Error("failed to download", err)
		return
	}
}
