package integrationtest_test

import (
	"log"
	"testing"
	"time"

	it "github.com/elisasre/go-common/v2/integrationtest"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestMain(m *testing.M) {
	itr := it.NewIntegrationTestRunner(
		it.OptBase("../"),
		it.OptTarget("./cmd/golden-demo"),
		it.OptCoverDir(it.IntegrationTestCoverDir),
		it.OptCompose("docker-compose.yaml", it.ComposeUpOptions(tc.Wait(true))),
		it.OptWaitHTTPReady("http://127.0.0.1:8080/healthz", time.Second*10),
		it.OptTestMain(m),
	)
	if err := itr.InitAndRun(); err != nil {
		log.Fatal(err)
	}
}
