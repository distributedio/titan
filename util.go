package titan

import (
	"fmt"
	"sync/atomic"

	"github.com/distributedio/titan/context"
	"github.com/twinj/uuid"

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

// PrintVersionInfo prints the server version info
func PrintVersionInfo() {
	fmt.Println("Welcome to Titan.")
	fmt.Println("Release Version: ", context.ReleaseVersion)
	fmt.Println("Git Commit Hash: ", context.GitHash)
	fmt.Println("Git Commit Log: ", context.GitLog)
	fmt.Println("Git Branch: ", context.GitBranch)
	fmt.Println("UTC Build Time:  ", context.BuildTS)
	fmt.Println("Golang compiler Version: ", context.GolangVersion)
}

//GetClientID starts with 1 and allocates clientID incrementally
func GetClientID() func() int64 {
	var id int64 = 1
	return func() int64 {
		return atomic.AddInt64(&id, 1)
	}
}

//GenerateTraceID grenerates a traceid for once a request
func GenerateTraceID() string { return uuid.NewV4().String() }
