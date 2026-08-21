package main

import (
	"encoding/json"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginabi"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginapi"
)

func TestRegistrationAndInterception(t *testing.T) {
	registered := configurePlugin(t)
	if !registered.Capabilities.RequestInterceptor {
		t.Fatal("request interceptor capability was not registered")
	}
	if registered.Metadata.Name != pluginID || registered.Metadata.Version != pluginVersion {
		t.Fatalf("unexpected metadata: %#v", registered.Metadata)
	}
	if registered.Metadata.GitHubRepository != "https://github.com/AstroQore/cpa-plugin-codex-responses-lite" {
		t.Fatalf("unexpected repository: %q", registered.Metadata.GitHubRepository)
	}

	for _, method := range []string{
		pluginabi.MethodRequestInterceptBefore,
		pluginabi.MethodRequestInterceptAfter,
	} {
		t.Run(method, func(t *testing.T) {
			matched := intercept(t, method, pluginapi.RequestInterceptRequest{
				RequestedModel: "opencode/grok-4.5",
			})
			if got := matched.Headers.Get(responsesLiteHeader); got != "true" {
				t.Fatalf("Responses Lite header = %q in %#v", got, matched.Headers)
			}

			unmatched := intercept(t, method, pluginapi.RequestInterceptRequest{
				RequestedModel: "opencode/other-model",
			})
			if len(unmatched.Headers) != 0 {
				t.Fatalf("unexpected headers for unmatched model: %#v", unmatched.Headers)
			}
		})
	}
}

func TestInterceptFallsBackToCurrentModel(t *testing.T) {
	configurePlugin(t)
	response := intercept(t, pluginabi.MethodRequestInterceptAfter, pluginapi.RequestInterceptRequest{
		Model: "opencode/grok-4.5",
	})
	if got := response.Headers.Get(responsesLiteHeader); got != "true" {
		t.Fatalf("Responses Lite header = %q in %#v", got, response.Headers)
	}
}

func configurePlugin(t *testing.T) registration {
	t.Helper()
	registerRequest, err := json.Marshal(lifecycleRequest{ConfigYAML: []byte(`
rules:
  - provider_prefix: opencode
    models:
      - grok-4.5
`)})
	if err != nil {
		t.Fatal(err)
	}
	return callAndDecode[registration](t, pluginabi.MethodPluginRegister, registerRequest)
}

func intercept(t *testing.T, method string, request pluginapi.RequestInterceptRequest) pluginapi.RequestInterceptResponse {
	t.Helper()
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return callAndDecode[pluginapi.RequestInterceptResponse](t, method, raw)
}

func callAndDecode[T any](t *testing.T, method string, request []byte) T {
	t.Helper()
	raw, err := handleMethod(method, request)
	if err != nil {
		t.Fatalf("handleMethod(%q) error = %v", method, err)
	}
	var wrapped envelope
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if !wrapped.OK {
		t.Fatalf("plugin returned error: %#v", wrapped.Error)
	}
	var result T
	if err := json.Unmarshal(wrapped.Result, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return result
}
