package autotest

import (
	"os"
	"testing"
	"time"

	"github.com/meitu/titan/tools/integration"
)

var (
	at *AutoClient
	an *Abnormal
)

func TestMain(m *testing.M) {
	integration.SetAuth("titan")
	go integration.Start()
	time.Sleep(time.Second)
	at = NewAutoClient()
	an = NewAbnormal()
	//TODO
	at.Start(integration.ServerAddr)
	an.Start(integration.ServerAddr)
	// Pool = newPool(integration.ServerAddr)
	v := m.Run()
	an.Close()
	integration.Close()
	at.Close()

	os.Exit(v)
}
