package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-dalle/v2/pkg/progress"
)

// startStatusPrinter launches a background goroutine that periodically queries the
// dalle progress manager and prints a concise aligned table of active runs.
// It intentionally writes directly to stderr so that normal stdout (if used) is
// not polluted and log rotation (which mirrors stderr) still captures status.
func startStatusPrinter(interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	// Print immediate startup notice + first status
	logger.Info(fmt.Sprintf("[status] reporter started (interval=%s)", interval))

	// Helper to build and output current status table; safe against panics
	printStatus := func() {
		defer func() {
			if r := recover(); r != nil { // never die silently
				logger.Info(fmt.Sprintf("[status] panic recovered: %v", r))
			}
		}()
		reports := progress.ActiveProgressReports()
		sort.Slice(reports, func(i, j int) bool {
			if reports[i].Series == reports[j].Series {
				return reports[i].Address < reports[j].Address
			}
			return reports[i].Series < reports[j].Series
		})
		type row struct {
			series, address, phase, pct, eta, elapsed, cache, err string
		}
		rows := []row{}
		for _, r := range reports {
			elapsed := ""
			if r.StartedNs > 0 {
				elapsed = time.Since(time.Unix(0, r.StartedNs)).Truncate(time.Second).String()
			}
			eta := ""
			if !r.Done && r.ETASeconds > 0 {
				eta = time.Duration(r.ETASeconds * float64(time.Second)).Truncate(time.Second).String()
			}
			rows = append(rows, row{
				series:  r.Series,
				address: shortened(r.Address),
				phase:   string(r.Current),
				pct:     fmt.Sprintf("%5.1f%%", r.Percent),
				eta:     eta,
				elapsed: elapsed,
				cache:   yesNo(r.CacheHit),
				err:     r.Error,
			})
		}
		// Determine widths dynamically from data only (no header rows printed)
		max := func(cur, val int) int {
			if val > cur {
				return val
			}
			return cur
		}
		w := struct{ series, address, phase, pct, eta, elapsed, cache, err int }{}
		for _, r := range rows {
			w.series = max(w.series, len(r.series))
			w.address = max(w.address, len(r.address))
			w.phase = max(w.phase, len(r.phase))
			w.pct = max(w.pct, len(r.pct))
			w.eta = max(w.eta, len(r.eta))
			w.elapsed = max(w.elapsed, len(r.elapsed))
			w.cache = max(w.cache, len(r.cache))
			w.err = max(w.err, len(r.err))
		}

		if len(rows) == 0 {
			logger.Info("(no active runs)")
		} else {
			if len(rows) > 1 {
				logger.Info("")
			}
			for _, r := range rows {
				logger.Info(fmt.Sprintf(colors.BrightYellow+"%-*s  %-*s  %-*s  %*s  %-*s  %-*s  %-*s  %-*s"+colors.Off,
					w.phase, r.phase,
					w.address, r.address,
					w.pct, r.pct,
					w.eta, r.eta,
					w.elapsed, r.elapsed,
					w.series, r.series,
					w.cache, r.cache,
					w.err, r.err,
				))
			}
		}
	}

	printStatus()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			printStatus()
		}
	}()
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// shortened returns a shortened address (0x + 4 + .. + 4) for readability.
func shortened(addr string) string {
	if len(addr) <= 12 { // already short
		return addr
	}
	return addr[:6] + "…" + addr[len(addr)-4:]
}
