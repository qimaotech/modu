package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qimaotech/modu/internal/core"
)

func (f *Formatter) FormatCheckoutResponse(feature string, results []core.CheckoutResult, operationErr error) string {
	if f.format == "json" {
		if results == nil {
			results = []core.CheckoutResult{}
		}
		response := struct {
			Success bool                  `json:"success"`
			Action  string                `json:"action"`
			Feature string                `json:"feature"`
			Results []core.CheckoutResult `json:"results"`
			Error   string                `json:"error,omitempty"`
		}{Success: operationErr == nil, Action: "checkout", Feature: feature, Results: results}
		if operationErr != nil {
			response.Error = operationErr.Error()
		}
		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Sprintf("{\"success\":false,\"error\":%q}\n", err.Error())
		}
		return string(data) + "\n"
	}
	var text strings.Builder
	if operationErr == nil {
		fmt.Fprintf(&text, "✓ 已接手 feature: %s\n", feature)
	} else {
		fmt.Fprintf(&text, "✗ 接手未完成: %s\n", feature)
	}
	for _, result := range results {
		fmt.Fprintf(&text, "  %s [%s] %s\n", result.Module, result.Status, result.Message)
		if result.Path != "" {
			fmt.Fprintf(&text, "    %s · %s\n    %s\n", result.Branch, result.Commit, result.Path)
		}
	}
	if operationErr != nil {
		fmt.Fprintf(&text, "%s\n", operationErr)
	}
	return text.String()
}
