package thanos

import (
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
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
	zap.L().Info("Welcome to Titan.")
	zap.L().Info("Server info", zap.String("Release Version", ReleaseVersion))
	zap.L().Info("Server info", zap.String("Git Commit Hash", GitHash))
	zap.L().Info("Server info", zap.String("Git Commit Log", GitLog))
	zap.L().Info("Server info", zap.String("Git Branch", GitBranch))
	zap.L().Info("Server info", zap.String("UTC Build Time", BuildTS))
	zap.L().Info("Server info", zap.String("Golang compiler Version", GolangVersion))
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
