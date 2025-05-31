package goignore

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// this file was adapted from the go-gitignore package:
// https://github.com/sabhiram/go-gitignore/blob/525f6e181f062064d83887ed2530e3b1ba0bc95a/ignore_ported_test.go

/*
The MIT License (MIT)

Copyright (c) 2015 Shaba Abhiram

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Validate the correct handling of the negation operator "!"
func TestCompileIgnoreLines_HandleIncludePattern(t *testing.T) {
	object := CompileIgnoreLines([]string{
		"/*",
		"!/foo",
		"/foo/*",
		"!/foo/bar",
	})

	assert.Equal(t, true, object.Match("a"), "a should match")
	assert.Equal(t, true, object.Match("foo/baz"), "foo/baz should match")
	assert.Equal(t, false, object.Match("foo"), "foo should not match")
	assert.Equal(t, false, object.Match("/foo/bar"), "/foo/bar should not match")
}

// Validate the correct handling of leading / chars
func TestCompileIgnoreLines_HandleLeadingSlash(t *testing.T) {
	object := CompileIgnoreLines([]string{
		"/a/b/c",
		"d/e/f",
		"/g",
	})

	assert.Equal(t, true, object.Match("a/b/c"), "a/b/c should match")
	assert.Equal(t, true, object.Match("a/b/c/d"), "a/b/c/d should match")
	assert.Equal(t, true, object.Match("d/e/f"), "d/e/f should match")
	assert.Equal(t, true, object.Match("g"), "g should match")
}

// Validate the correct handling of files starting with # or !
func TestCompileIgnoreLines_HandleLeadingSpecialChars(t *testing.T) {
	object := CompileIgnoreLines([]string{
		"# Comment",
		"\\#file.txt",
		"\\!file.txt",
		"file.txt",
	})

	assert.Equal(t, true, object.Match("#file.txt"), "#file.txt should match")
	assert.Equal(t, true, object.Match("!file.txt"), "!file.txt should match")
	assert.Equal(t, true, object.Match("a/!file.txt"), "a/!file.txt should match")
	assert.Equal(t, true, object.Match("file.txt"), "file.txt should match")
	assert.Equal(t, true, object.Match("a/file.txt"), "a/file.txt should match")
	assert.Equal(t, false, object.Match("file2.txt"), "file2.txt should not match")

}

// Validate the correct handling matching files only within a given folder
func TestCompileIgnoreLines_HandleAllFilesInDir(t *testing.T) {
	object := CompileIgnoreLines([]string{"Documentation/*.html"})

	assert.Equal(t, true, object.Match("Documentation/git.html"), "Documentation/git.html should match")
	assert.Equal(t, false, object.Match("Documentation/ppc/ppc.html"), "Documentation/ppc/ppc.html should not match")
	assert.Equal(t, false, object.Match("tools/perf/Documentation/perf.html"), "tools/perf/Documentation/perf.html should not match")
}

// Validate the correct handling of "**"
func TestCompileIgnoreLines_HandleDoubleStar(t *testing.T) {
	object := CompileIgnoreLines([]string{"**/foo", "bar"})

	assert.Equal(t, true, object.Match("foo"), "foo should match")
	assert.Equal(t, true, object.Match("baz/foo"), "baz/foo should match")
	assert.Equal(t, true, object.Match("bar"), "bar should match")
	assert.Equal(t, true, object.Match("baz/bar"), "baz/bar should match")
}

// Validate the correct handling of leading slash
func TestCompileIgnoreLines_HandleLeadingSlashPath(t *testing.T) {
	object := CompileIgnoreLines([]string{"/*.c"})

	assert.Equal(t, true, object.Match("hello.c"), "hello.c should match")
	assert.Equal(t, false, object.Match("foo/hello.c"), "foo/hello.c should not match")
}

func ExampleCompileIgnoreLines() {
	ignoreObject := CompileIgnoreLines([]string{"node_modules", "*.out", "foo/*.c"})

	// You can test the ignoreObject against various paths using the
	// "MatchesPath()" interface method. This pretty much is up to
	// the users interpretation. In the case of a ".gitignore" file,
	// a "match" would indicate that a given path would be ignored.
	fmt.Println(ignoreObject.Match("node_modules/test/foo.js"))
	fmt.Println(ignoreObject.Match("node_modules2/test.out"))
	fmt.Println(ignoreObject.Match("test/foo.js"))

	// Output:
	// true
	// true
	// false
}

func TestCompileIgnoreLines_CheckNestedDotFiles(t *testing.T) {
	lines := []string{
		"**/external/**/*.md",
		"**/external/**/*.json",
		"**/external/**/*.gzip",
		"**/external/**/.*ignore",

		"**/external/foobar/*.css",
		"**/external/barfoo/less",
		"**/external/barfoo/scss",
	}
	object := CompileIgnoreLines(lines)
	assert.NotNil(t, object, "returned object should not be nil")

	assert.Equal(t, true, object.Match("external/foobar/angular.foo.css"), "external/foobar/angular.foo.css")
	assert.Equal(t, true, object.Match("external/barfoo/.gitignore"), "external/barfoo/.gitignore")
	assert.Equal(t, true, object.Match("external/barfoo/.bower.json"), "external/barfoo/.bower.json")
}

func TestCompileIgnoreLines_CarriageReturn(t *testing.T) {
	lines := []string{"abc/def\r", "a/b/c\r", "b\r"}
	object := CompileIgnoreLines(lines)

	assert.Equal(t, true, object.Match("abc/def/child"), "abc/def/child should match")
	assert.Equal(t, true, object.Match("a/b/c/d"), "a/b/c/d should match")

	assert.Equal(t, false, object.Match("abc"), "abc should not match")
	assert.Equal(t, false, object.Match("def"), "def should not match")
	assert.Equal(t, false, object.Match("bd"), "bd should not match")
}

func TestCompileIgnoreLines_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	lines := []string{"abc/def", "a/b/c", "b"}
	object := CompileIgnoreLines(lines)

	assert.Equal(t, true, object.Match("abc\\def\\child"), "abc\\def\\child should match")
	assert.Equal(t, true, object.Match("a\\b\\c\\d"), "a\\b\\c\\d should match")
}

func TestWildCardFiles(t *testing.T) {
	gitIgnore := []string{"*.swp", "/foo/*.wat", "bar/*.txt"}
	object := CompileIgnoreLines(gitIgnore)

	// Paths which are targeted by the above "lines"
	assert.Equal(t, true, object.Match("yo.swp"), "should ignore all swp files")
	assert.Equal(t, true, object.Match("something/else/but/it/hasyo.swp"), "should ignore all swp files in other directories")

	assert.Equal(t, true, object.Match("foo/bar.wat"), "should ignore all wat files in foo - nonpreceding /")
	assert.Equal(t, true, object.Match("/foo/something.wat"), "should ignore all wat files in foo - preceding /")

	assert.Equal(t, true, object.Match("bar/something.txt"), "should ignore all txt files in bar - nonpreceding /")
	assert.Equal(t, true, object.Match("/bar/somethingelse.txt"), "should ignore all txt files in bar - preceding /")

	// Paths which are not targeted by the above "lines"
	assert.Equal(t, false, object.Match("something/not/infoo/wat.wat"), "wat files should only be ignored in foo")
	assert.Equal(t, false, object.Match("something/not/infoo/wat.txt"), "txt files should only be ignored in bar")
}

func TestPrecedingSlash(t *testing.T) {
	gitIgnore := []string{"/foo", "bar/"}
	object := CompileIgnoreLines(gitIgnore)

	assert.Equal(t, true, object.Match("foo/bar.wat"), "should ignore all files in foo - nonpreceding /")
	assert.Equal(t, true, object.Match("/foo/something.txt"), "should ignore all files in foo - preceding /")

	assert.Equal(t, true, object.Match("bar/something.txt"), "should ignore all files in bar - nonpreceding /")
	assert.Equal(t, true, object.Match("/bar/somethingelse.go"), "should ignore all files in bar - preceding /")
	assert.Equal(t, true, object.Match("/boo/something/bar/boo.txt"), "should block all files if bar is a sub directory")

	assert.Equal(t, false, object.Match("something/foo/something.txt"), "should only ignore top level foo directories- not nested")
}

func TestDirOnlyMatching(t *testing.T) {
	gitIgnore := []string{"foo/", "bar/"}
	object := CompileIgnoreLines(gitIgnore)

	assert.Equal(t, true, object.Match("foo/"), "should match foo directory")
	assert.Equal(t, true, object.Match("bar/"), "should match bar directory")
	assert.Equal(t, false, object.Match("foo"), "should not match foo file")
	assert.Equal(t, false, object.Match("bar"), "should not match bar file")
	assert.Equal(t, true, object.Match("foo/bar"), "should match nested files in foo")
	assert.Equal(t, true, object.Match("bar/foo"), "should match nested files in bar")
}
