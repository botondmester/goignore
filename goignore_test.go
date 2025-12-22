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

func ExampleCompileIgnoreLines() {
	ignoreObject := CompileIgnoreLines([]string{"node_modules", "*.out", "foo/*.c"})

	// You can test the ignoreObject against various paths using the
	// "Match()" interface method. This pretty much is up to
	// the users interpretation. In the case of a ".gitignore" file,
	// a "match" would indicate that a given path would be ignored.
	fmt.Println(ignoreObject.MatchesPath("node_modules/test/foo.js"))
	fmt.Println(ignoreObject.MatchesPath("node_modules2/test.out"))
	fmt.Println(ignoreObject.MatchesPath("test/foo.js"))

	// Output:
	// true
	// true
	// false
}

func ExampleCompileIgnoreFile() {
	ignoreObject, err := CompileIgnoreFile(".gitignore")

	// err returns an error from os.ReadFile()
	if err != nil {
		fmt.Println("Error reading .gitignore file:", err)
		return
	}

	// You can test the ignoreObject against various paths using the
	// "Match()" interface method.
	// int this example, we test paths against the .gitignore file of this package.
	fmt.Println(ignoreObject.MatchesPath("bin/goignore.so"))
	fmt.Println(ignoreObject.MatchesPath("goignore.test"))
	fmt.Println(ignoreObject.MatchesPath("go.mod"))

	// Output:
	// true
	// true
	// false
}

// Validate the correct handling of the negation operator "!"
func TestCompileIgnoreLines_HandleIncludePattern(t *testing.T) {
	ignoreObject := CompileIgnoreLines([]string{
		"/*",
		"!/foo",
		"/foo/*",
		"!/foo/bar",
	})

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("a"), "a should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("foo/baz"), "foo/baz should match")
	assert.Equal(t, false, ignoreObject.MatchesPath("foo"), "foo should not match")
	assert.Equal(t, false, ignoreObject.MatchesPath("/foo/bar"), "/foo/bar should not match")
}

// Validate the correct handling of leading / chars
func TestCompileIgnoreLines_HandleLeadingSlash(t *testing.T) {
	ignoreObject := CompileIgnoreLines([]string{
		"/a/b/c",
		"d/e/f",
		"/g",
	})

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("a/b/c"), "a/b/c should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("a/b/c/d"), "a/b/c/d should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("d/e/f"), "d/e/f should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("g"), "g should match")
}

// Validate the correct handling of files starting with # or !
func TestCompileIgnoreLines_HandleLeadingSpecialChars(t *testing.T) {
	ignoreObject := CompileIgnoreLines([]string{
		"# Comment",
		"\\#file.txt",
		"\\!file.txt",
		"file.txt",
	})

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("#file.txt"), "#file.txt should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("!file.txt"), "!file.txt should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("a/!file.txt"), "a/!file.txt should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("file.txt"), "file.txt should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("a/file.txt"), "a/file.txt should match")
	assert.Equal(t, false, ignoreObject.MatchesPath("file2.txt"), "file2.txt should not match")

}

// Validate the correct handling matching files only within a given folder
func TestCompileIgnoreLines_HandleAllFilesInDir(t *testing.T) {
	ignoreObject := CompileIgnoreLines([]string{"Documentation/*.html"})

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("Documentation/git.html"), "Documentation/git.html should match")
	assert.Equal(t, false, ignoreObject.MatchesPath("Documentation/ppc/ppc.html"), "Documentation/ppc/ppc.html should not match")
	assert.Equal(t, false, ignoreObject.MatchesPath("tools/perf/Documentation/perf.html"), "tools/perf/Documentation/perf.html should not match")
}

// Validate the correct handling of "**"
func TestCompileIgnoreLines_HandleDoubleStar(t *testing.T) {
	ignoreObject := CompileIgnoreLines([]string{"**/foo", "bar"})

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("foo"), "foo should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("baz/foo"), "baz/foo should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("bar"), "bar should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("baz/bar"), "baz/bar should match")
}

// Validate the correct handling of leading slash
func TestCompileIgnoreLines_HandleLeadingSlashPath(t *testing.T) {
	object := CompileIgnoreLines([]string{"/*.c"})

	assert.NotNil(t, object, "Returned object should not be nil")

	assert.Equal(t, true, object.MatchesPath("hello.c"), "hello.c should match")
	assert.Equal(t, false, object.MatchesPath("foo/hello.c"), "foo/hello.c should not match")
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
	ignoreObject := CompileIgnoreLines(lines)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("external/foobar/angular.foo.css"), "external/foobar/angular.foo.css should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("external/barfoo/.gitignore"), "external/barfoo/.gitignore should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("external/barfoo/.bower.json"), "external/barfoo/.bower.json should match")
}

func TestCompileIgnoreLines_CarriageReturn(t *testing.T) {
	lines := []string{"abc/def\r", "a/b/c\r", "b\r"}
	ignoreObject := CompileIgnoreLines(lines)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("abc/def/child"), "abc/def/child should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("a/b/c/d"), "a/b/c/d should match")

	assert.Equal(t, false, ignoreObject.MatchesPath("abc"), "abc should not match")
	assert.Equal(t, false, ignoreObject.MatchesPath("def"), "def should not match")
	assert.Equal(t, false, ignoreObject.MatchesPath("bd"), "bd should not match")
}

func TestCompileIgnoreLines_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	lines := []string{"abc/def", "a/b/c", "b"}
	ignoreObject := CompileIgnoreLines(lines)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("abc\\def\\child"), "abc\\def\\child should match")
	assert.Equal(t, true, ignoreObject.MatchesPath("a\\b\\c\\d"), "a\\b\\c\\d should match")
}

func TestWildCardFiles(t *testing.T) {
	gitIgnore := []string{"*.swp", "/foo/*.wat", "bar/*.txt"}
	ignoreObject := CompileIgnoreLines(gitIgnore)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	// Paths which are targeted by the above "lines"
	assert.Equal(t, true, ignoreObject.MatchesPath("yo.swp"), "should ignore all swp files")
	assert.Equal(t, true, ignoreObject.MatchesPath("something/else/but/it/hasyo.swp"), "should ignore all swp files in other directories")

	assert.Equal(t, true, ignoreObject.MatchesPath("foo/bar.wat"), "should ignore all wat files in foo - nonpreceding /")
	assert.Equal(t, true, ignoreObject.MatchesPath("/foo/something.wat"), "should ignore all wat files in foo - preceding /")

	assert.Equal(t, true, ignoreObject.MatchesPath("bar/something.txt"), "should ignore all txt files in bar - nonpreceding /")
	assert.Equal(t, true, ignoreObject.MatchesPath("/bar/somethingelse.txt"), "should ignore all txt files in bar - preceding /")

	// Paths which are not targeted by the above "lines"
	assert.Equal(t, false, ignoreObject.MatchesPath("something/not/infoo/wat.wat"), "wat files should only be ignored in foo")
	assert.Equal(t, false, ignoreObject.MatchesPath("something/not/infoo/wat.txt"), "txt files should only be ignored in bar")
}

func TestPrecedingSlash(t *testing.T) {
	gitIgnore := []string{"/foo", "bar/"}
	ignoreObject := CompileIgnoreLines(gitIgnore)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("foo/bar.wat"), "should ignore all files in foo - nonpreceding /")
	assert.Equal(t, true, ignoreObject.MatchesPath("/foo/something.txt"), "should ignore all files in foo - preceding /")

	assert.Equal(t, true, ignoreObject.MatchesPath("bar/something.txt"), "should ignore all files in bar - nonpreceding /")
	assert.Equal(t, true, ignoreObject.MatchesPath("/bar/somethingelse.go"), "should ignore all files in bar - preceding /")
	assert.Equal(t, true, ignoreObject.MatchesPath("/boo/something/bar/boo.txt"), "should block all files if bar is a sub directory")

	assert.Equal(t, false, ignoreObject.MatchesPath("something/foo/something.txt"), "should only ignore top level foo directories - not nested")
}

func TestDirOnlyMatching(t *testing.T) {
	gitIgnore := []string{"foo/", "bar/"}
	ignoreObject := CompileIgnoreLines(gitIgnore)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("foo/"), "should match foo directory")
	assert.Equal(t, true, ignoreObject.MatchesPath("bar/"), "should match bar directory")
	assert.Equal(t, false, ignoreObject.MatchesPath("foo"), "should not match foo file")
	assert.Equal(t, false, ignoreObject.MatchesPath("bar"), "should not match bar file")
	assert.Equal(t, true, ignoreObject.MatchesPath("foo/bar"), "should match nested files in foo")
	assert.Equal(t, true, ignoreObject.MatchesPath("bar/foo"), "should match nested files in bar")
}

func TestCharacterClasses(t *testing.T) {
	gitIgnore := []string{"[a-zA-Z*!]-files"}
	ignoreObject := CompileIgnoreLines(gitIgnore)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, true, ignoreObject.MatchesPath("a-files"), "should match a-files")
	assert.Equal(t, true, ignoreObject.MatchesPath("g-files"), "should match g-files")
	assert.Equal(t, true, ignoreObject.MatchesPath("z-files"), "should match z-files")
	assert.Equal(t, true, ignoreObject.MatchesPath("!-files"), "should match !-files")
	assert.Equal(t, true, ignoreObject.MatchesPath("*-files"), "should match *-files")
	assert.Equal(t, false, ignoreObject.MatchesPath("8-files"), "should not match 8-files")

	gitIgnore = []string{"[!a-zA-Z*!]-files"}
	ignoreObject = CompileIgnoreLines(gitIgnore)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")

	assert.Equal(t, false, ignoreObject.MatchesPath("a-files"), "should not match a-files")
	assert.Equal(t, false, ignoreObject.MatchesPath("g-files"), "should not match g-files")
	assert.Equal(t, false, ignoreObject.MatchesPath("z-files"), "should not match z-files")
	assert.Equal(t, false, ignoreObject.MatchesPath("!-files"), "should not match !-files")
	assert.Equal(t, false, ignoreObject.MatchesPath("*-files"), "should not match *-files")
	assert.Equal(t, true, ignoreObject.MatchesPath("8-files"), "should match 8-files")
}

func TestUnclosedCharacterClass(t *testing.T) {
	gitIgnore := []string{"[a-z][A-Z*-files"}
	ignoreObject := CompileIgnoreLines(gitIgnore)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")
	assert.Equal(t, true, ignoreObject.MatchesPath("a[A-Z-files"), "should match a[A-Z-files")

	gitIgnore = []string{"[a-z][[]A-Z*-files"}
	ignoreObject = CompileIgnoreLines(gitIgnore)

	assert.NotNil(t, ignoreObject, "Returned object should not be nil")
	assert.Equal(t, true, ignoreObject.MatchesPath("a[A-Z-files"), "should match a[A-Z-files")
}
