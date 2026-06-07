// Package hello is a self-registering example workflow.
//
// Importing it for its side effects (`_ "…/examples/hello"`) makes the "hello"
// workflow discoverable through registry.Default. It shows the whole model in
// miniature: a tool, a skill, and a command wired into a DAG, with data
// threaded through the shared Values bag.
package hello

import (
	"context"
	"fmt"
	"strings"

	"github.com/emil-jacero/mycelium/component"
	"github.com/emil-jacero/mycelium/registry"
)

func init() {
	registry.Register("hello", build)
}

func build() (component.Component, error) {
	// fetch (tool): produces raw input.
	fetch := component.NewTool("fetch", func(_ context.Context, _ component.Values) (component.Values, error) {
		return component.Values{"raw": "the quick brown fox"}, nil
	})

	// summarize (skill): depends on fetch, transforms its output.
	summarize := component.NewSkill("summarize", func(_ context.Context, in component.Values) (component.Values, error) {
		raw, _ := in["raw"].(string)
		words := strings.Fields(raw)
		return component.Values{"summary": fmt.Sprintf("%d words", len(words))}, nil
	}, "fetch")

	// report (command): depends on summarize, renders the final result.
	report := component.NewCommand("report", func(_ context.Context, in component.Values) (component.Values, error) {
		summary, _ := in["summary"].(string)
		return component.Values{"report": "summary: " + summary}, nil
	}, "summarize")

	return component.NewWorkflow("hello", fetch, summarize, report)
}
