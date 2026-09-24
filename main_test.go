package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginabi"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginapi"
)

func TestRegistrationAndInterception(t *testing.T) {
	registered := configurePlugin(t)
	if !registered.Capabilities["request_interceptor"] || !registered.Capabilities["management_api"] {
		t.Fatalf("unexpected capabilities: %#v", registered.Capabilities)
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

func TestManagementWizard(t *testing.T) {
	reg := callAndDecode[map[string]any](t, pluginabi.MethodManagementRegister, nil)
	resources, ok := reg["resources"].([]any)
	if !ok || len(resources) != 1 {
		t.Fatalf("unexpected management registration: %#v", reg)
	}
	page, ok := resources[0].(map[string]any)
	if !ok || page["path"] != "/config-wizard" {
		t.Fatalf("unexpected resource entry: %#v", resources[0])
	}

	raw, err := json.Marshal(managementRPCRequest{Method: "GET", Path: "/plugins/aq-codex-responses-lite/config-wizard"})
	if err != nil {
		t.Fatal(err)
	}
	var wrapped envelope
	var out []byte
	out, err = handleMethod(pluginabi.MethodManagementHandle, raw)
	if err != nil {
		t.Fatalf("management handle GET error: %v", err)
	}
	if err := json.Unmarshal(out, &wrapped); err != nil || !wrapped.OK {
		t.Fatalf("management handle GET failed: %v %#v", err, wrapped.Error)
	}
	var pageResp managementRPCResponse
	if err := json.Unmarshal(wrapped.Result, &pageResp); err != nil {
		t.Fatalf("decode management response: %v", err)
	}
	if pageResp.StatusCode != 200 || len(pageResp.Body) == 0 {
		t.Fatalf("unexpected page response: %#v", pageResp)
	}

	post, err := json.Marshal(managementRPCRequest{Method: "POST"})
	if err != nil {
		t.Fatal(err)
	}
	out, _ = handleMethod(pluginabi.MethodManagementHandle, post)
	if err := json.Unmarshal(out, &wrapped); err != nil || !wrapped.OK {
		t.Fatalf("management handle POST failed: %v %#v", err, wrapped.Error)
	}
	if err := json.Unmarshal(wrapped.Result, &pageResp); err != nil {
		t.Fatalf("decode management response: %v", err)
	}
	if pageResp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for POST, got %#v", pageResp)
	}
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
