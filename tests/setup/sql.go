package setup

import (
	"strings"

	"github.com/grantfbarnes/card-judge/static"
	"github.com/grantfbarnes/card-judge/tests/util"
)

// getSQLFileList returns the ordered list of SQL files to execute.
// Uses the shared list from static package, stripping the "sql/" prefix
// since test setup reads from src/static/sql/ base path.
func getSQLFileList() []string {
	files := make([]string, len(static.SQLFiles))
	for i, f := range static.SQLFiles {
		// Strip "sql/" prefix since database.go adds the full base path
		files[i] = strings.TrimPrefix(f, "sql/")
	}
	return files
}

// getSQLBasePath finds the SQL directory relative to current working directory
func getSQLBasePath() string {
	return util.FindPath(
		util.SQLDir,
		"./"+util.SQLDir,
		"../"+util.SQLDir,
		"../../"+util.SQLDir,
		"../../../"+util.SQLDir,
	)
}
