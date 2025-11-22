package titan

import (
	"fmt"
	"sync/atomic"

	"github.com/distributedio/titan/context"
	"github.com/twinj/uuid"
)

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
