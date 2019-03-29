package volume

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/puppetlabs/wash/plugin"
	"github.com/stretchr/testify/assert"
)

// Generated with
// `docker run --rm -it -v=/test/fixture:/mnt busybox find /mnt/ -mindepth 1 -exec stat -c '%s %X %Y %Z %f %n' {} \;`
const mountpoint = "mnt"
const fixture = `
96 1550611510 1550611448 1550611448 41ed mnt/path
96 1550611510 1550611448 1550611448 41ed mnt/path/has
96 1550611510 1550611448 1550611448 41ed mnt/path/has/got
96 1550611510 1550611458 1550611458 41ed mnt/path/has/got/some
0 1550611458 1550611458 1550611458 81a4 mnt/path/has/got/some/legs
96 1550611510 1550611453 1550611453 41ed mnt/path1
0 1550611453 1550611453 1550611453 81a4 mnt/path1/a file
96 1550611510 1550611441 1550611441 41ed mnt/path2
64 1550611510 1550611441 1550611441 41ed mnt/path2/dir
`

func TestStatParse(t *testing.T) {
	actualAttr, path, err := StatParse("96 1550611510 1550611448 1550611448 41ed mnt/path")
	assert.Nil(t, err)
	assert.Equal(t, "mnt/path", path)
	expectedAttr := plugin.EntryAttributes{}
	expectedAttr.
		SetAtime(time.Unix(1550611510, 0)).
		SetMtime(time.Unix(1550611448, 0)).
		SetCtime(time.Unix(1550611448, 0)).
		SetMode(0755 | os.ModeDir).
		SetSize(96)
	assert.Equal(t, expectedAttr, actualAttr)

	actualAttr, path, err = StatParse("0 1550611458 1550611458 1550611458 81a4 mnt/path/has/got/some/legs")
	assert.Nil(t, err)
	assert.Equal(t, "mnt/path/has/got/some/legs", path)
	expectedAttr = plugin.EntryAttributes{}
	expectedAttr.
		SetAtime(time.Unix(1550611458, 0)).
		SetMtime(time.Unix(1550611458, 0)).
		SetCtime(time.Unix(1550611458, 0)).
		SetMode(0644).
		SetSize(0)
	assert.Equal(t, expectedAttr, actualAttr)

	_, _, err = StatParse("stat: failed")
	assert.Equal(t, errors.New("Stat did not return 6 components: stat: failed"), err)

	_, _, err = StatParse("-1 1550611510 1550611448 1550611448 41ed mnt/path")
	if assert.NotNil(t, err) {
		assert.Equal(t, &strconv.NumError{Func: "ParseUint", Num: "-1", Err: strconv.ErrSyntax}, err)
	}

	_, _, err = StatParse("0 2019-01-01 2019-01-01 2019-01-01 41ed mnt/path")
	if assert.NotNil(t, err) {
		assert.Equal(t, &strconv.NumError{Func: "ParseInt", Num: "2019-01-01", Err: strconv.ErrSyntax}, err)
	}

	_, _, err = StatParse("96 1550611510 1550611448 1550611448 zebra mnt/path")
	if assert.NotNil(t, err) {
		assert.Regexp(t, regexp.MustCompile("parse.*mode.*zebra"), err.Error())
	}
}

func TestStatParseAll(t *testing.T) {
	dmap, err := StatParseAll(strings.NewReader(fixture), mountpoint)
	assert.Nil(t, err)
	assert.NotNil(t, dmap)
	assert.Equal(t, 8, len(dmap))
	for _, dir := range []string{"", "/path", "/path/has", "/path/has/got", "/path/has/got/some", "/path1", "/path2", "/path2/dir"} {
		assert.NotNil(t, dmap[dir])
	}
	for _, file := range []string{"/path/has/got/some/legs", "/path1/a file"} {
		assert.Nil(t, dmap[file])
	}

	for _, node := range []string{"/path", "/path1", "/path2"} {
		assert.NotNil(t, dmap[""][node])
	}

	expectedAttr := plugin.EntryAttributes{}
	expectedAttr.
		SetAtime(time.Unix(1550611453, 0)).
		SetMtime(time.Unix(1550611453, 0)).
		SetCtime(time.Unix(1550611453, 0)).
		SetMode(0644).
		SetSize(0)
	assert.Equal(t, expectedAttr, dmap["/path1"]["a file"])

	expectedAttr = plugin.EntryAttributes{}
	expectedAttr.
		SetAtime(time.Unix(1550611510, 0)).
		SetMtime(time.Unix(1550611441, 0)).
		SetCtime(time.Unix(1550611441, 0)).
		SetMode(0755 | os.ModeDir).
		SetSize(64)
	assert.Equal(t, expectedAttr, dmap["/path2"]["dir"])

	expectedAttr = plugin.EntryAttributes{}
	expectedAttr.
		SetAtime(time.Unix(1550611510, 0)).
		SetMtime(time.Unix(1550611448, 0)).
		SetCtime(time.Unix(1550611448, 0)).
		SetMode(0755 | os.ModeDir).
		SetSize(96)
	assert.Equal(t, expectedAttr, dmap["/path"]["has"])
}
