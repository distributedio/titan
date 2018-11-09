package thanos

import (
	"fmt"
	"sync/atomic"

	"github.com/twinj/uuid"
	"gitlab.meitu.com/platform/thanos/context"

	"go.uber.org/zap"
)

func logVersionInfo() {
	zap.L().Info("Welcome to Titan.")
	zap.L().Info("Server info", zap.String("Release Version", context.ReleaseVersion))
	zap.L().Info("Server info", zap.String("Git Commit Hash", context.GitHash))
	zap.L().Info("Server info", zap.String("Git Commit Log", context.GitLog))
	zap.L().Info("Server info", zap.String("Git Branch", context.GitBranch))
	zap.L().Info("Server info", zap.String("UTC Build Time", context.BuildTS))
	zap.L().Info("Server info", zap.String("Golang compiler Version", context.GolangVersion))
}

// PrintVersionInfo print the server version info
func PrintVersionInfo() {
	fmt.Println("Welcome to Titan.")
	fmt.Println("Release Version: ", context.ReleaseVersion)
	fmt.Println("Git Commit Hash: ", context.GitHash)
	fmt.Println("Git Commit Log: ", context.GitLog)
	fmt.Println("Git Branch: ", context.GitBranch)
	fmt.Println("UTC Build Time:  ", context.BuildTS)
	fmt.Println("Golang compiler Version: ", context.GolangVersion)
}

func GetClientID() func() int64 {
	var id int64 = 1
	return func() int64 {
		return atomic.AddInt64(&id, 1)
	}
}

func GenerateTraceID() string { return uuid.NewV4().String() }
