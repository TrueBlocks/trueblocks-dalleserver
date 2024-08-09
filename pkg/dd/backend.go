package dd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/TrueBlocks/trueblocks-core/sdk/v3"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-dalleserver/pkg/openai"
	"github.com/joho/godotenv"
)

type Maker struct {
	authorTemplate *template.Template
	promptTemplate *template.Template
	dataTemplate   *template.Template
	terseTemplate  *template.Template
	titleTemplate  *template.Template
	Series         Series `json:"series"`
	apiKeys        map[string]string
	databases      map[string][]string
	dalleCache     map[string]*DalleDress
	LastSeries     string
}

func NewMaker(series string) *Maker {
	maker := Maker{
		dalleCache: make(map[string]*DalleDress),
		apiKeys:    make(map[string]string),
		databases:  make(map[string][]string),
		LastSeries: series,
	}
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	} else if maker.apiKeys["openAi"] = os.Getenv("OPENAI_API_KEY"); maker.apiKeys["openAi"] == "" {
		log.Fatal("No OPENAI_API_KEY key found")
	}

	// Initialize your data here
	var err error
	if maker.promptTemplate, err = template.New("prompt").Parse(promptTemplate); err != nil {
		logger.Fatal("could not create prompt template:", err)
	}
	if maker.dataTemplate, err = template.New("data").Parse(dataTemplate); err != nil {
		logger.Fatal("could not create data template:", err)
	}
	if maker.titleTemplate, err = template.New("terse").Parse(titleTemplate); err != nil {
		logger.Fatal("could not create title template:", err)
	}
	if maker.terseTemplate, err = template.New("terse").Parse(terseTemplate); err != nil {
		logger.Fatal("could not create terse template:", err)
	}
	if maker.authorTemplate, err = template.New("author").Parse(authorTemplate); err != nil {
		logger.Fatal("could not create prompt template:", err)
	}
	logger.Info()
	logger.Info("Compiled templates")

	maker.ReloadDatabases()

	return &maker
}

var dalleCacheMutex sync.Mutex

func (maker *Maker) MakeDalleDress(addressIn string) (*DalleDress, error) {
	dalleCacheMutex.Lock()
	defer dalleCacheMutex.Unlock()
	if maker.dalleCache[addressIn] != nil {
		logger.Info("Returning cached dalle for", addressIn)
		return maker.dalleCache[addressIn], nil
	}

	address := addressIn
	logger.Info("Making dalle for", addressIn)
	if strings.HasSuffix(address, ".eth") {
		opts := sdk.NamesOptions{
			Terms: []string{address},
		}
		if names, _, err := opts.Names(); err != nil {
			return nil, fmt.Errorf("error getting names for %s", address)
		} else {
			if len(names) > 0 {
				address = names[0].Address.Hex()
			}
		}
	}
	logger.Info("Resolved", addressIn)

	parts := strings.Split(address, ",")
	seed := parts[0] + reverse(parts[0])
	if len(seed) < 66 {
		return nil, fmt.Errorf("seed length is less than 66")
	}
	if strings.HasPrefix(seed, "0x") {
		seed = seed[2:66]
	}

	fn := validFilename(address)
	if maker.dalleCache[fn] != nil {
		logger.Info("Returning cached dalle for", addressIn)
		return maker.dalleCache[fn], nil
	}

	dd := DalleDress{
		Original:  addressIn,
		Filename:  fn,
		Seed:      seed,
		AttribMap: make(map[string]Attribute),
	}

	for i := 0; i < len(dd.Seed); i = i + 8 {
		index := len(dd.Attribs)
		attr := NewAttribute(maker.databases, index, dd.Seed[i:i+6])
		dd.Attribs = append(dd.Attribs, attr)
		dd.AttribMap[attr.Name] = attr
		if i+4+6 < len(dd.Seed) {
			index = len(dd.Attribs)
			attr = NewAttribute(maker.databases, index, dd.Seed[i+4:i+4+6])
			dd.Attribs = append(dd.Attribs, attr)
			dd.AttribMap[attr.Name] = attr
		}
	}

	suff := maker.Series.Suffix
	dd.DataPrompt, _ = dd.ExecuteTemplate(maker.dataTemplate, nil)
	dd.ReportOn(addressIn, filepath.Join(suff, "data"), "txt", dd.DataPrompt)
	dd.TitlePrompt, _ = dd.ExecuteTemplate(maker.titleTemplate, nil)
	dd.ReportOn(addressIn, filepath.Join(suff, "title"), "txt", dd.TitlePrompt)
	dd.TersePrompt, _ = dd.ExecuteTemplate(maker.terseTemplate, nil)
	dd.ReportOn(addressIn, filepath.Join(suff, "terse"), "txt", dd.TersePrompt)
	dd.Prompt, _ = dd.ExecuteTemplate(maker.promptTemplate, nil)
	dd.ReportOn(addressIn, filepath.Join(suff, "prompt"), "txt", dd.Prompt)
	fn = filepath.Join("output", maker.Series.Suffix, "enhanced", dd.Filename+".txt")
	dd.EnhancedPrompt = ""
	if file.FileExists(fn) {
		dd.EnhancedPrompt = file.AsciiFileToString(fn)
	}

	maker.dalleCache[dd.Filename] = &dd
	maker.dalleCache[addressIn] = &dd

	return &dd, nil
}

func (maker *Maker) GetAppSeries(addr string) string {
	return maker.Series.String()
}

func (maker *Maker) GetJson(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		return dd.String()
	}
}

func (maker *Maker) GetData(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		return dd.DataPrompt
	}
}

func (maker *Maker) GetTitle(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		return dd.TitlePrompt
	}
}

func (maker *Maker) GetTerse(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		return dd.TersePrompt
	}
}

func (maker *Maker) GetPrompt(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		return dd.Prompt
	}
}

func (maker *Maker) GetEnhanced(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		return dd.EnhancedPrompt
	}
}

func (maker *Maker) GetFilename(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		return dd.Filename
	}
}

func (maker *Maker) Save(addr string) bool {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return false
	} else {
		dd.ReportOn(addr, filepath.Join(maker.Series.Suffix, "selector"), "json", dd.String())
		return true
	}
}

func (maker *Maker) GenerateEnhanced(addr string) string {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error()
	} else {
		authorType, _ := dd.ExecuteTemplate(maker.authorTemplate, nil)
		if dd.EnhancedPrompt, err = openai.EnhancePrompt(maker.GetPrompt(addr), authorType); err != nil {
			logger.Fatal(err.Error())
		}
		msg := " DO NOT PUT TEXT IN THE IMAGE. "
		dd.EnhancedPrompt = msg + dd.EnhancedPrompt + msg
		return dd.EnhancedPrompt
	}
}

func (maker *Maker) GenerateImage(addr string) (string, error) {
	if dd, err := maker.MakeDalleDress(addr); err != nil {
		return err.Error(), err
	} else {
		suff := maker.Series.Suffix
		dd.EnhancedPrompt = maker.GenerateEnhanced(addr)
		dd.ReportOn(addr, filepath.Join(suff, "enhanced"), "txt", dd.EnhancedPrompt)
		_ = maker.Save(addr)
		imageData := openai.ImageData{
			TitlePrompt:    dd.TitlePrompt,
			TersePrompt:    dd.TersePrompt,
			EnhancedPrompt: dd.EnhancedPrompt,
			SeriesName:     maker.Series.Suffix,
			Filename:       dd.Filename,
		}
		if err := openai.RequestImage(&imageData); err != nil {
			return err.Error(), err
		}
		return dd.EnhancedPrompt, nil
	}
}

// validFilename returns a valid filename from the input string
func validFilename(in string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		in = strings.ReplaceAll(in, char, "_")
	}
	in = strings.TrimSpace(in)
	in = strings.ReplaceAll(in, "__", "_")
	return in
}

// reverse returns the reverse of the input string
func reverse(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

func (maker *Maker) ReloadDatabases() {
	maker.Series = Series{}
	maker.databases = make(map[string][]string)

	var err error
	if maker.Series, err = maker.LoadSeries(); err != nil {
		logger.Fatal(err)
	}
	logger.Info("Loaded series:", maker.Series.Suffix)

	for _, db := range DatabaseNames {
		if maker.databases[db] == nil {
			if lines, err := maker.toLines(db); err != nil {
				logger.Fatal(err)
			} else {
				maker.databases[db] = lines
				for i := 0; i < len(maker.databases[db]); i++ {
					maker.databases[db][i] = strings.Replace(maker.databases[db][i], "v0.1.0,", "", -1)
				}
			}
		}
	}
	logger.Info("Loaded", len(DatabaseNames), "databases")
}

func (maker *Maker) LoadSeries() (Series, error) {
	lastSeries := maker.LastSeries // "simple" // maker.GetSession().LastSeries
	if lastSeries == "" {
		lastSeries = "simple"
	}
	fn := filepath.Join("./output/series", lastSeries+".json")
	str := strings.TrimSpace(file.AsciiFileToString(fn))
	logger.Info("lastSeries", lastSeries)
	if len(str) == 0 || !file.FileExists(fn) {
		logger.Info("No series found, creating a new one", fn)
		ret := Series{
			Suffix: "simple",
		}
		ret.SaveSeries(fn, 0)
		return ret, nil
	}

	bytes := []byte(str)
	var s Series
	if err := json.Unmarshal(bytes, &s); err != nil {
		logger.Error("could not unmarshal series:", err)
		return Series{}, err
	}

	s.Suffix = strings.Trim(strings.ReplaceAll(s.Suffix, " ", "-"), "-")
	s.SaveSeries(filepath.Join("./output/series", s.Suffix+".json"), 0)
	return s, nil
}

func (maker *Maker) toLines(db string) ([]string, error) {
	filename := "./databases/" + db + ".csv"
	lines := file.AsciiFileToLines(filename)
	lines = lines[1:] // skip header
	var err error
	if len(lines) == 0 {
		err = fmt.Errorf("could not load %s", filename)
	} else {
		fn := strings.ToUpper(db[:1]) + db[1:]
		if filter, err := maker.Series.GetFilter(fn); err != nil {
			return lines, err

		} else {
			if len(filter) == 0 {
				return lines, nil
			}

			filtered := make([]string, 0, len(lines))
			for _, line := range lines {
				for _, f := range filter {
					if strings.Contains(line, f) {
						filtered = append(filtered, line)
					}
				}
			}
			lines = filtered
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "none")
	}

	return lines, err
}
