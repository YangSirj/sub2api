package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Parallel Claude Code agents share a root session. Both Anthropic API-key
// forwarding paths must preserve their identity for the downstream scheduler.
func TestGatewayClaudeAgentHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, passthrough := range []bool{false, true} {
		name := "standard"
		if passthrough {
			name = "passthrough"
		}
		t.Run(name, func(t *testing.T) {
			for _, agent := range []string{"", "agent-a", "agent-b"} {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
				c.Request.Header.Set("X-Claude-Code-Session-Id", "root-session")
				if agent != "" {
					c.Request.Header.Set("X-Claude-Code-Agent-Id", agent)
					c.Request.Header.Set("X-Claude-Code-Parent-Agent-Id", "parent-agent")
				}
				c.Request.Header.Set("Authorization", "Bearer client-secret")
				c.Request.Header.Set("Cookie", "client-cookie=secret")
				account := &Account{ID: 201, Platform: PlatformAnthropic, Type: AccountTypeAPIKey}
				body := []byte(`{"model":"claude-sonnet-4-5","max_tokens":16,"messages":[{"role":"user","content":"hello"}]}`)
				svc := &GatewayService{cfg: &config.Config{}}
				var req *http.Request
				var err error
				if passthrough {
					req, _, err = svc.buildUpstreamRequestAnthropicAPIKeyPassthrough(context.Background(), c, account, body, "upstream-key")
				} else {
					req, _, err = svc.buildUpstreamRequest(context.Background(), c, account, body, "upstream-key", "apikey", "claude-sonnet-4-5", false, false)
				}
				require.NoError(t, err)
				require.Equal(t, "root-session", getHeaderRaw(req.Header, "x-claude-code-session-id"))
				require.Equal(t, agent, getHeaderRaw(req.Header, "x-claude-code-agent-id"))
				parent := ""
				if agent != "" {
					parent = "parent-agent"
				}
				require.Equal(t, parent, getHeaderRaw(req.Header, "x-claude-code-parent-agent-id"))
				require.Equal(t, "upstream-key", getHeaderRaw(req.Header, "x-api-key"))
				require.Empty(t, getHeaderRaw(req.Header, "authorization"))
				require.Empty(t, getHeaderRaw(req.Header, "cookie"))
			}
		})
	}
}
