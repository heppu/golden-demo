package integrationtest_test

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/elisasre/go-common/v2/golden"
	it "github.com/elisasre/go-common/v2/integrationtest"
	"github.com/elisasre/go-common/v2/must"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestMain(m *testing.M) {
	os.Setenv("GOFLAGS", "-tags=integration")
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

func TestAPI(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		body         io.Reader
		expectedCode int
	}{
		{
			name:   "create task",
			method: "POST",
			path:   "/api/v1/tasks",
			body: strings.NewReader(`
				{
					"title": "testing 2",
					"description": "asdasd",
					"status": "waiting"
				}
			`),
			expectedCode: 200,
		},
		{
			name:   "modify task",
			method: "PUT",
			path:   "/api/v1/tasks/1",
			body: strings.NewReader(`
				{
					"title": "testing 2",
					"description": "asdasd",
					"status": "working"
				}
			`),
			expectedCode: 200,
		},
		{
			name:         "list tasks",
			method:       "GET",
			path:         "/api/v1/tasks",
			expectedCode: 200,
		},
		{
			name:         "delete tasks",
			method:       "DELETE",
			path:         "/api/v1/tasks/1",
			expectedCode: 204,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := must.NewRequest(t, tt.method, "http://127.0.0.1:8080"+tt.path, tt.body)
			golden.Request(t, http.DefaultClient, req, tt.expectedCode)
		})
	}
}
