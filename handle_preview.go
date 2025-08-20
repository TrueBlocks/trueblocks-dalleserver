package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:0;padding:1.2rem;background:#111;color:#eee}
h1{margin-top:0;font-size:1.4rem}
header{display:flex;align-items:center;gap:1rem;margin-bottom:1rem}
.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(180px,1fr));gap:12px}
figure{margin:0;background:#1e1e1e;padding:8px;border-radius:8px;display:flex;flex-direction:column;gap:6px;box-shadow:0 2px 4px rgba(0,0,0,.4)}
.figure-img-wrapper{position:relative;width:100%;padding-top:100%;background:#222;border-radius:4px;overflow:hidden}
.figure-img-wrapper img{position:absolute;inset:0;width:100%;height:100%;object-fit:contain}
figcaption{font-size:.65rem;line-height:1.1em;word-break:break-word}
footer{margin-top:2rem;font-size:.65rem;color:#888;text-align:center}
a{color:#7fb0ff;text-decoration:none}
a:hover{text-decoration:underline}
.series-group{margin-bottom:2rem}
.series-group h2{font-size:1rem;margin:0 0 .5rem 0;border-bottom:1px solid #333;padding-bottom:4px}
button{background:#333;color:#eee;border:1px solid #444;padding:4px 10px;border-radius:4px;cursor:pointer}
button:hover{background:#3d3d3d}
</style>
<script>
function filterSeries(){const q=document.getElementById('filter').value.toLowerCase();document.querySelectorAll('.series-group').forEach(g=>{const s=g.dataset.series;g.style.display=s.includes(q)?'block':'none';});}
</script>
</head><body>
<header>
  <h1>Annotated Image Preview</h1>
  <input id="filter" placeholder="Filter series..." oninput="filterSeries()" style="padding:6px;border-radius:4px;border:1px solid #333;background:#222;color:#eee" />
  <a href="/" style="margin-left:auto">Home</a>
</header>
{{if .Images}}
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
{{else}}
  <p>No annotated images found yet. Trigger generation via /dalle/&lt;series&gt;/&lt;address&gt;?generate=1</p>
{{end}}
<footer>Generated at {{.Now.Format "2006-01-02 15:04:05 MST"}} • Found {{len .Images}} images.</footer>
</body></html>`))

func (a *App) handlePreview(w http.ResponseWriter, r *http.Request) {
	root := "output"
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
		http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
	}
}
