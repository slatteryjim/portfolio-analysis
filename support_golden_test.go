package portfolio_analysis

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

// Drop this file in any package you want to support "Golden Files" in tests.
//
// Explained here:
//   https://www.youtube.com/watch?v=8hQG7QlcLBk&feature=youtu.be&t=736
//   https://ieftimov.com/post/testing-in-go-golden-files/
//
// Typically you would add a Makefile target as well, for example:
//     # Runs `go test` with the -update flag to update the "golden files"
//     # to reflect the output of the latest code changes.
//     test-update-goldenfiles:
//     	rm -rf testdata/*  # might need to delve into subfolders later
//     	go test -update ./...
//
// So you can run `make test-update-goldenfiles` to update the ".golden" file used to compare output.

var _update = flag.Bool("update", false, "update .golden files of the tests")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// ExpectMatchesGoldenFile verifies that the actualContent matches the golden file
// that exists for this specific test case, t.Name().
// If the "-update" flag was set, the golden file will be updated to match the
// given content.
func ExpectMatchesGoldenFile(t *testing.T, actualContent string) {
	t.Helper()
	g := NewGomegaWithT(t)

	// create subfolder if needed
	const dirname = "testdata"
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		g.Expect(os.Mkdir(dirname, 0755)).To(Succeed(), "creating directory")
	}
	basename := t.Name()
	basename = strings.ReplaceAll(basename, "/", "__") // don't use subfolders
	golden := filepath.Join(dirname, basename+".golden")
	if *_update {
		err := ioutil.WriteFile(golden, []byte(actualContent), 0644)
		g.Expect(err).To(Succeed(), "writing golden file")
	}
	byts, err := ioutil.ReadFile(golden)
	g.Expect(err).ToNot(HaveOccurred(), "reading golden file")
	expected := string(byts)

	if expected != actualContent {
		t.Errorf("Content did not match golden file. (Run `make test-update-goldenfiles` if you need to update the golden file to match the latest code changes.)")
		fmt.Println("Golden content:")
		fmt.Println(expected)
		fmt.Println()
		fmt.Println("Got:")
		fmt.Println(actualContent)
	}
}
