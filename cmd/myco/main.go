// Command myco is the Mycelium CLI.
//
// This POC is stdlib-only (no external deps) so it builds offline. Subcommands:
//
//	myco version            print the version
//	myco list               list registered components/workflows
//	myco graph <id>         print a workflow's execution plan (DAG topo order)
//	myco run <id>           run a workflow on the local substrate
package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/emil-jacero/mycelium/component"
	"github.com/emil-jacero/mycelium/registry"
	"github.com/emil-jacero/mycelium/runtime"

	// Plugins self-register via init(). Adding a workflow to the CLI is a
	// one-line blank import — the pluggability seam in action.
	_ "github.com/emil-jacero/mycelium/examples/hello"
)

const version = "0.0.1-poc"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "version":
		fmt.Println("myco", version)
	case "list":
		err = cmdList()
	case "graph":
		err = cmdGraph(os.Args[2:])
	case "run":
		err = cmdRun(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`myco — compose tools, skills, scripts & commands into agentic workflows

usage:
  myco version          print version
  myco list             list registered components
  myco graph <id>       print a workflow's execution plan
  myco run <id>         run a workflow
`)
}

func cmdList() error {
	ids := registry.Default.IDs()
	if len(ids) == 0 {
		fmt.Println("no components registered")
		return nil
	}
	for _, id := range ids {
		c, err := registry.Default.Build(id)
		if err != nil {
			return err
		}
		fmt.Printf("%-12s %s\n", c.Kind(), id)
	}
	return nil
}

func cmdGraph(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: myco graph <id>")
	}
	c, err := registry.Default.Build(args[0])
	if err != nil {
		return err
	}
	wf, ok := c.(*component.Workflow)
	if !ok {
		return fmt.Errorf("%q is a %s, not a workflow", args[0], c.Kind())
	}
	plan, err := wf.Plan()
	if err != nil {
		return err
	}
	fmt.Printf("plan for %q (%d steps):\n", args[0], len(plan))
	for i, id := range plan {
		step, _ := wf.Node(id)
		deps := step.Requires()
		sort.Strings(deps)
		if len(deps) == 0 {
			fmt.Printf("  %d. %-10s %s\n", i+1, step.Kind(), id)
		} else {
			fmt.Printf("  %d. %-10s %s  <- %v\n", i+1, step.Kind(), id, deps)
		}
	}
	return nil
}

func cmdRun(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: myco run <id>")
	}
	c, err := registry.Default.Build(args[0])
	if err != nil {
		return err
	}

	// Headless listener: print each step as it runs. A TUI would swap this
	// for a renderer — same events, different surface.
	sub := runtime.New("local", func(e runtime.Event) {
		switch e.Phase {
		case runtime.PhaseStart:
			fmt.Printf("→ %-10s %s\n", e.Kind, e.ID)
		case runtime.PhaseError:
			fmt.Printf("✗ %-10s %s: %v\n", e.Kind, e.ID, e.Err)
		}
	})

	out, err := sub.Run(context.Background(), c, component.Values{})
	if err != nil {
		return err
	}

	fmt.Println("\nresult:")
	keys := make([]string, 0, len(out))
	for k := range out {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %s = %v\n", k, out[k])
	}
	return nil
}
