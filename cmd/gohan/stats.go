package main

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/bmf-san/gohan/internal/model"
)

// phaseTimer accumulates wall-clock durations for named build phases. It is
// safe to call Phase repeatedly for the same name; durations are summed.
type phaseTimer struct {
	order    []string
	duration map[string]time.Duration
}

func newPhaseTimer() *phaseTimer {
	return &phaseTimer{duration: make(map[string]time.Duration)}
}

// Phase runs fn while recording its wall-clock duration under name. The first
// time a name is seen its order is remembered so the final report is stable.
func (p *phaseTimer) Phase(name string, fn func() error) error {
	if _, seen := p.duration[name]; !seen {
		p.order = append(p.order, name)
	}
	t0 := time.Now()
	err := fn()
	p.duration[name] += time.Since(t0)
	return err
}

// writeStats writes a human-readable phase-timing report to w. Phases are
// printed in the order they were first observed, followed by the total.
func (p *phaseTimer) writeStats(w io.Writer, total time.Duration) {
	_, _ = fmt.Fprintln(w, "stats:")
	for _, name := range p.order {
		_, _ = fmt.Fprintf(w, "  %-12s %v\n", name+":", p.duration[name].Round(time.Microsecond))
	}
	_, _ = fmt.Fprintf(w, "  %-12s %v\n", "total:", total.Round(time.Millisecond))
}

// writeExplain reports the reason for the current build's scope of work to w.
// `forceFull` indicates that a full rebuild was forced (CLI flag, missing
// manifest, or config hash change). `changeSet` lists incremental changes
// when a partial rebuild was performed.
func writeExplain(w io.Writer, forceFull bool, fullReason string, changeSet *model.ChangeSet) {
	_, _ = fmt.Fprintln(w, "explain:")
	if forceFull {
		if fullReason == "" {
			fullReason = "full build forced"
		}
		_, _ = fmt.Fprintf(w, "  full build: %s\n", fullReason)
		return
	}
	if changeSet == nil {
		_, _ = fmt.Fprintln(w, "  no change set available")
		return
	}
	if len(changeSet.AddedFiles) == 0 && len(changeSet.ModifiedFiles) == 0 && len(changeSet.DeletedFiles) == 0 {
		_, _ = fmt.Fprintln(w, "  no content changes detected")
		return
	}
	printList(w, "added", changeSet.AddedFiles)
	printList(w, "modified", changeSet.ModifiedFiles)
	printList(w, "deleted", changeSet.DeletedFiles)
}

func printList(w io.Writer, label string, files []string) {
	if len(files) == 0 {
		return
	}
	sorted := append([]string(nil), files...)
	sort.Strings(sorted)
	_, _ = fmt.Fprintf(w, "  %s (%d):\n", label, len(sorted))
	for _, f := range sorted {
		_, _ = fmt.Fprintf(w, "    %s\n", f)
	}
}
