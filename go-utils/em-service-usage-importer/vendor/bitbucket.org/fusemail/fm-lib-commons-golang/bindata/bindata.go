/*
Package bindata wraps go-bindata and provides files & templates utilities.
Distinguishes whether sys built or not, using bindata if built, and direct file access otherwise.
Better than bindata's debug stubs: https://github.com/jteeuwen/go-bindata#debug-vs-release-builds
because it automatically handles new files without any further generation,
in addition to other common functionality across packages.

TODO (when necessary) - create funcs to access BinAssetDir and BinAssetNames.
*/
package bindata

import (
	htmltpl "html/template"
	"io/ioutil"
	texttpl "text/template"

	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
	log "github.com/sirupsen/logrus"
)

// BinAsset* holds bindata's funcs, injected via SetupBin.
var (
	BinAsset      func(string) ([]byte, error)
	BinAssetDir   func(string) ([]string, error)
	BinAssetNames func() []string
)

// Setup injects bindata's funcs into utils package.
func Setup(
	fnAsset func(string) ([]byte, error),
	fnAssetDir func(string) ([]string, error),
	fnAssetNames func() []string,
) {
	BinAsset = fnAsset
	BinAssetDir = fnAssetDir
	BinAssetNames = fnAssetNames
}

// LoadFile reads and returns file bytes, with nil or error.
func LoadFile(filename string) ([]byte, error) {
	built := sys.Built()
	fn := BinAsset
	if !built {
		fn = ioutil.ReadFile
	}
	byts, err := fn(filename)
	if err != nil {
		return nil, err
	}
	return byts, nil
}

// RequireFile does the same as LoadFile plus panics if any error.
func RequireFile(filename string) []byte {
	byts, err := LoadFile(filename)
	if err != nil {
		log.WithFields(log.Fields{"filename": filename, "err": err}).Panic("Failed to load file")
	}
	return byts
}

// RequireTextTemplate reads and parses file as text, and returns template.
// Panics if any error.
func RequireTextTemplate(filename string) *texttpl.Template {
	str := string(RequireFile(filename))
	return texttpl.Must(texttpl.New(filename).Parse(str))
}

// RequireHTMLTemplate reads and parses file as HTML, and returns template.
// Panics if any error.
func RequireHTMLTemplate(filename string) *htmltpl.Template {
	str := string(RequireFile(filename))
	return htmltpl.Must(htmltpl.New(filename).Parse(str))
}
