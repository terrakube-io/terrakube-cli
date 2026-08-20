package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	// Version is populated at build time via -ldflags.
	Version = "dev"
	// Commit is populated at build time via -ldflags.
	Commit = "none"
	// Date is populated at build time via -ldflags.
	Date = "unknown"
)

// BuildInfo encapsulates version and runtime build metadata.
type BuildInfo struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
	Platform  string
}

// GetBuildInfo resolves build information from ldflags variables or runtime debug info.
func GetBuildInfo() BuildInfo {
	info := BuildInfo{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		if info.Version == "dev" || info.Version == "" {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				info.Version = bi.Main.Version
			}
		}

		var vcsRev, vcsTime string
		var vcsModified bool
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				vcsRev = s.Value
			case "vcs.time":
				vcsTime = s.Value
			case "vcs.modified":
				vcsModified = s.Value == "true"
			}
		}

		if (info.Commit == "none" || info.Commit == "") && vcsRev != "" {
			if len(vcsRev) > 7 {
				info.Commit = vcsRev[:7]
			} else {
				info.Commit = vcsRev
			}
			if vcsModified {
				info.Commit += "+dirty"
			}
		}

		if (info.Date == "unknown" || info.Date == "") && vcsTime != "" {
			info.Date = vcsTime
		}
	}

	return info
}

// FormatVersion returns the multi-line structured version output.
func FormatVersion() string {
	info := GetBuildInfo()
	return fmt.Sprintf("Version:    %s\nGit Commit: %s\nBuilt At:   %s\nGo Version: %s\nOS/Arch:    %s",
		info.Version,
		info.Commit,
		info.Date,
		info.GoVersion,
		info.Platform,
	)
}
