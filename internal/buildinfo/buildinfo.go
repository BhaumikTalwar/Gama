package buildinfo

import "fmt"

var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

func BuildInfo() string {
	return fmt.Sprintf("Build Info : Gama\n Version: %s\nCommit: %s\nBuild Time: %s", Version, Commit, BuildTime)
}

func VersionStr() string {
	return Version
}

func CommitStr() string {
	return Commit
}
