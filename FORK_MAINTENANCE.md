# vm2api integration patch

This fork preserves Claude Code's `x-claude-code-agent-id` and
`x-claude-code-parent-agent-id` through the existing Anthropic request-header
allowlist. Independent agents share a root session, so dropping these headers
causes a downstream scheduler to serialize unrelated agent requests.

`TestGatewayClaudeAgentHeaders` exercises both standard API-key forwarding and
API-key passthrough. It checks root, sibling, and nested-agent identity, and
verifies that inbound credentials and cookies are still removed.

Branch maintenance:

- `main`: upstream synchronization.
- `vm2api-integration`: integration changes on current upstream.
- `production/agent-headers-v0.2.17`: minimal backport based on production's
  `060d0160417782cd962a0c7f460f6cd4565d4916`, without bundling an upstream upgrade.

Before shipping a newer upstream version, merge upstream into the integration
branch, retain or confirm the equivalent header behavior, and run:

```sh
cd backend
go test ./internal/service -run 'TestGatewayClaudeAgentHeaders|TestGatewayService_AnthropicAPIKeyPassthrough' -count=1
```

Deploy the binary built from the reviewed fork commit. The stock updater still
downloads upstream releases; using it directly would replace this patched
binary. Creating a fork alone does not change the updater's release source.

The companion vm2api change gives each Claude Code agent its own seat and VM
affinity. Header forwarding alone does not remove vm2api's root-family VM lock.
