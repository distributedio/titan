package thanos

import (
	"fmt"
	"sync/atomic"

	log "gitlab.meitu.com/gocommons/logbunny"
)

// Version information.
var (
	ReleaseVersion = "None"
	BuildTS        = "None"
	GitHash        = "None"
	GitBranch      = "None"
	GitLog         = "None"
	GolangVersion  = "None"
	ConfigFile     = "None"
)

func logVersionInfo() {
	log.Info("Welcome to Titan.")
	log.Info("Server info", log.String("Release Version", ReleaseVersion))
	log.Info("Server info", log.String("Git Commit Hash", GitHash))
	log.Info("Server info", log.String("Git Commit Log", GitLog))
	log.Info("Server info", log.String("Git Branch", GitBranch))
	log.Info("Server info", log.String("UTC Build Time", BuildTS))
	log.Info("Server info", log.String("Golang compiler Version", GolangVersion))
}

// PrintVersionInfo print the server version info
func PrintVersionInfo() {
	fmt.Println("Welcome to Titan.")
	fmt.Println("Release Version: ", ReleaseVersion)
	fmt.Println("Git Commit Hash: ", GitHash)
	fmt.Println("Git Commit Log: ", GitLog)
	fmt.Println("Git Branch: ", GitBranch)
	fmt.Println("UTC Build Time:  ", BuildTS)
	fmt.Println("Golang compiler Version: ", GolangVersion)
}

func GetClientID() uint64 {
	var id uint64 = 1
	return func() uint64 {
		return atomic.AddUint64(&id, 1)
	}()
}
