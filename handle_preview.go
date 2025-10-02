package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TrueBlocks/trueblocks-dalle/v2/pkg/storage"
)

// imageMeta holds minimal data for gallery rendering.
type imageMeta struct {
	Series  string
	Address string
	Path    string // relative path for <img src>
	ModTime time.Time
}

var previewTpl = template.Must(template.New("preview").Parse(`<!DOCTYPE html>
<html><head><meta charset="utf-8" />
<title>DalleServer Preview</title>
<style>
:root {
  --grid-columns: 4;
}
body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:0;padding:0;background:#111;color:#eee;padding-top:80px}
h1{margin-top:0;font-size:1.4rem}
header{position:fixed;top:0;left:0;right:0;z-index:1000;background:#111;border-bottom:1px solid #333;display:flex;align-items:center;gap:1rem;padding:1.2rem;flex-wrap:wrap}
.controls{display:flex;align-items:center;gap:1rem;flex-wrap:wrap}
.grid-control{display:flex;align-items:center;gap:0.5rem}
.grid-control label{font-size:0.8rem;color:#ccc}
.grid-control input[type="range"]{width:120px}
.grid-control span{font-size:0.75rem;color:#999;min-width:15px}
.grid{display:grid;grid-template-columns:repeat(var(--grid-columns),1fr);gap:12px}
figure{margin:0;background:#1e1e1e;padding:8px;border-radius:8px;display:flex;flex-direction:column;gap:6px;box-shadow:0 2px 4px rgba(0,0,0,.4)}
.figure-img-wrapper{position:relative;width:100%;padding-top:100%;background:#222;border-radius:4px;overflow:hidden}
.figure-img-wrapper img{position:absolute;inset:0;width:100%;height:100%;object-fit:contain}
figcaption{font-size:.65rem;line-height:1.1em;word-break:break-word}
footer{margin-top:2rem;font-size:.65rem;color:#888;text-align:center}
a{color:#7fb0ff;text-decoration:none}
a:hover{text-decoration:underline}
.series-group{margin-bottom:2rem}
.series-group h2{position:sticky;top:80px;z-index:100;background:#111;font-size:1rem;margin:0 0 .5rem 0;border-bottom:1px solid #333;padding:8px 1.2rem;margin-left:-1.2rem;margin-right:-1.2rem}
button{background:#333;color:#eee;border:1px solid #444;padding:4px 10px;border-radius:4px;cursor:pointer}
button:hover{background:#3d3d3d}
input[type="range"]{background:#222;border:1px solid #444;border-radius:4px}
input[type="text"]{padding:6px;border-radius:4px;border:1px solid #333;background:#222;color:#eee}
</style>
<script>
function filterSeries(){const q=document.getElementById('filter').value.toLowerCase();document.querySelectorAll('.series-group').forEach(g=>{const s=g.dataset.series;g.style.display=s.includes(q)?'block':'none';});}

function updateGridColumns() {
  const slider = document.getElementById('gridColumns');
  const display = document.getElementById('gridValue');
  const columns = slider.value;
  display.textContent = columns;
  document.documentElement.style.setProperty('--grid-columns', columns);
  // Save to localStorage
  localStorage.setItem('previewGridColumns', columns);
}

function initializeGridControl() {
  const slider = document.getElementById('gridColumns');
  const display = document.getElementById('gridValue');
  
  // Load saved value from localStorage, default to 4
  const savedColumns = localStorage.getItem('previewGridColumns') || '4';
  slider.value = savedColumns;
  display.textContent = savedColumns;
  document.documentElement.style.setProperty('--grid-columns', savedColumns);
  
  // Add event listener
  slider.addEventListener('input', updateGridColumns);
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', initializeGridControl);
</script>
</head><body>
<header>
  <h1>Annotated Image Preview</h1>
  <div class="controls">
    <input id="filter" placeholder="Filter series..." oninput="filterSeries()" style="padding:6px;border-radius:4px;border:1px solid #333;background:#222;color:#eee" />
    <div class="grid-control">
      <label for="gridColumns">Columns:</label>
      <input type="range" id="gridColumns" min="2" max="8" value="4" />
      <span id="gridValue">4</span>
    </div>
  </div>
  <a href="/" style="margin-left:auto">Home</a>
</header>
{{if .Images}}
  <main style="padding:0 1.2rem">
  {{range $series, $list := .BySeries}}
    <section class="series-group" data-series="{{$series}}">
      <h2>{{$series}} ({{len $list}})</h2>
      <div class="grid">
        {{range $list}}
						<figure>
							<div class="figure-img-wrapper">
								<img loading="lazy" src="/files/{{.Path}}" alt="{{.Series}} {{.Address}}" />
							</div>
							<figcaption>{{.Address}}<br/><span style="color:#666">{{.ModTime.Format "2006-01-02 15:04:05"}}</span></figcaption>
						</figure>
        {{end}}
      </div>
    </section>
  {{end}}
  </main>
{{else}}
  <main style="padding:0 1.2rem">
  <p>No annotated images found yet. Trigger generation via /dalle/&lt;series&gt;/&lt;address&gt;?generate=1</p>
  </main>
{{end}}
<footer style="padding:0 1.2rem;margin-top:2rem;font-size:.65rem;color:#888;text-align:center">Generated at {{.Now.Format "2006-01-02 15:04:05 MST"}} • Found {{len .Images}} images.</footer>
</body></html>`))

func (a *App) handlePreview(w http.ResponseWriter, r *http.Request) {
	root := storage.OutputDir()
	var images []imageMeta
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".png") {
			return nil
		}
		if !strings.Contains(path, string(filepath.Separator)+"annotated"+string(filepath.Separator)) {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) < 3 {
			return nil
		}
		series := parts[0]
		addressFile := parts[len(parts)-1]
		address := strings.TrimSuffix(addressFile, filepath.Ext(addressFile))
		info, statErr := os.Stat(path)
		if statErr != nil {
			return nil
		}
		images = append(images, imageMeta{Series: series, Address: address, Path: rel, ModTime: info.ModTime()})
		return nil
	})
	bySeries := map[string][]imageMeta{}
	for _, im := range images {
		bySeries[im.Series] = append(bySeries[im.Series], im)
	}
	for k := range bySeries {
		list := bySeries[k]
		sort.Slice(list, func(i, j int) bool { return list[i].ModTime.After(list[j].ModTime) })
		bySeries[k] = list
	}
	keys := make([]string, 0, len(bySeries))
	for k := range bySeries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	data := struct {
		Images   []imageMeta
		BySeries map[string][]imageMeta
		Series   []string
		Now      time.Time
	}{Images: images, BySeries: bySeries, Series: keys, Now: time.Now()}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := previewTpl.Execute(w, data); err != nil {
		requestID := GenerateRequestID()
		apiErr := NewAPIError(
			ErrorTemplateExecution,
			"Template execution failed",
			err.Error(),
		).WithRequestID(requestID)
		WriteErrorResponse(w, apiErr, http.StatusInternalServerError)
	}
}
