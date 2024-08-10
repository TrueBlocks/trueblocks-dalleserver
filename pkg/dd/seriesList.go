package dd

import (
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/walk"
)

func SeriesList() []string {
	seriesList := []string{}
	vFunc := func(fn string, vP any) (bool, error) {
		if strings.HasSuffix(fn, ".json") {
			fn = strings.ReplaceAll(fn, "output/series/", "")
			fn = strings.ReplaceAll(fn, ".json", "")
			seriesList = append(seriesList, fn)
		}
		return true, nil
	}
	_ = walk.ForEveryFileInFolder("./output/series", vFunc, nil)
	return seriesList
}

func IsValidSeries(series string, validSeries []string) bool {
	if len(validSeries) == 0 {
		return true
	}

	for _, s := range validSeries {
		if s == series {
			return true
		}
	}

	return false
}
