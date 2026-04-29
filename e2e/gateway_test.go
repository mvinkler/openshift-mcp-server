//go:build e2e

package e2e

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/containers/kubernetes-mcp-server/internal/test"
	"github.com/stretchr/testify/suite"
)

type GatewaySuite struct {
	suite.Suite
	*test.McpClient
	endpoint string
}

func (s *GatewaySuite) SetupSuite() {
	endpoint := os.Getenv("MCP_GATEWAY_ENDPOINT")
	if endpoint == "" {
		s.T().Skip("MCP_GATEWAY_ENDPOINT not set, skipping e2e tests")
	}
	s.endpoint = endpoint
}

func (s *GatewaySuite) SetupTest() {
	s.McpClient = test.NewMcpClient(s.T(), nil,
		test.WithEndpoint(s.endpoint),
		test.WithTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // e2e test with self-signed certs
		}),
		test.WithClientInfo("e2e-gateway-test", "1.0.0"),
	)
}

func (s *GatewaySuite) TearDownTest() {
	if s.McpClient != nil {
		s.McpClient.Close()
	}
}

func (s *GatewaySuite) TestInitialize() {
	s.NotNil(s.InitializeResult, "MCP initialize should succeed")
	s.NotEmpty(s.InitializeResult.ServerInfo.Name, "server info should have a name")
}

func (s *GatewaySuite) TestToolsHaveKubePrefix() {
	result, err := s.ListTools()
	s.Require().NoError(err)
	s.Require().NotNil(result)
	s.Require().NotEmpty(result.Tools, "gateway should expose at least one tool")

	for _, tool := range result.Tools {
		s.True(strings.HasPrefix(tool.Name, "kube_"),
			"tool %q should have kube_ prefix", tool.Name)
	}
}

func TestGateway(t *testing.T) {
	suite.Run(t, new(GatewaySuite))
}
