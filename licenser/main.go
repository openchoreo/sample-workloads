// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Flag variables

var (
	flagCheckOnly = flag.Bool(
		"check-only",
		false,
		"Only verify headers (exit 1 if non-compliant)",
	)
	flagHolder = flag.String(
		"c",
		"",
		`Copyright holder, e.g. "The OpenChoreo Authors"`,
	)
)

// Constants
const licenseID = "Apache-2.0"

// Header detection / generation

var (
	reCopyright = regexp.MustCompile(`^(//|#) Copyright (\d{4}) (.+)$`)
	reSPDX      = regexp.MustCompile(`^(//|#) SPDX-License-Identifier: (Apache-2\.0)$`)
)

// File helpers

func isSourceFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".php", ".py", ".js", ".jsx", ".ts", ".tsx", ".bal", ".rb", ".java":
		return true
	default:
		return false
	}
}

func getCommentPrefix(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".php", ".js", ".jsx", ".ts", ".tsx", ".java":
		return "//"
	case ".py", ".rb":
		return "#"
	case ".bal":
		return "//"
	default:
		return "//" // default to // for unknown types
	}
}

func shortHeader(year, holder, commentPrefix string) string {
	return fmt.Sprintf(
		"%s Copyright %s %s\n%s SPDX-License-Identifier: %s",
		commentPrefix, year, holder, commentPrefix, licenseID,
	)
}
func hasValidHeader(path, holder string) (bool, error) {
	commentPrefix := getCommentPrefix(path)
	
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	var lines []string
	for scan.Scan() {
		line := scan.Text()

		// Skip any leading blank lines
		if strings.TrimSpace(line) == "" && len(lines) == 0 {
			continue
		}
		lines = append(lines, line)

		// We need three lines: copyright, SPDX, blank
		if len(lines) == 3 {
			break
		}
	}

	// Must have exactly the three expected lines
	if len(lines) < 3 {
		return false, nil
	}

	// Create dynamic regexes based on comment prefix
	reCopyright := regexp.MustCompile(fmt.Sprintf(`^%s Copyright (\d{4}) (.+)$`, regexp.QuoteMeta(commentPrefix)))
	reSPDX := regexp.MustCompile(fmt.Sprintf(`^%s SPDX-License-Identifier: (Apache-2\.0)$`, regexp.QuoteMeta(commentPrefix)))

	m1 := reCopyright.FindStringSubmatch(lines[0])
	m2 := reSPDX.FindStringSubmatch(lines[1])
	blank := strings.TrimSpace(lines[2]) == ""

	if m1 == nil || m2 == nil || !blank {
		return false, nil
	}

	return m1[2] == holder && m2[1] == licenseID, nil
}

func prependHeader(path, header string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, append([]byte(header+"\n\n"), src...), 0o644)
}

// Core processing loop

func process(path, header, holder string, fix bool) (changed bool, err error) {
	ok, err := hasValidHeader(path, holder)
	if err != nil || ok {
		return false, err
	}
	if !fix {
		return true, nil // non-compliant
	}
	return true, prependHeader(path, header)
}

func walk(root, holder string, fix bool) ([]string, error) {
	var nonCompliant []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isSourceFile(p) {
			return err
		}
		commentPrefix := getCommentPrefix(p)
		header := shortHeader(fmt.Sprint(time.Now().Year()), holder, commentPrefix)
		changed, err := process(p, header, holder, fix)
		if err != nil {
			return err
		}
		if changed {
			nonCompliant = append(nonCompliant, p)
		}
		return nil
	})
	return nonCompliant, err
}

// CLI

const usageText = `
Licenser is a tool to enforce short SPDX license headers in source files.

OVERVIEW
  licenser verifies that each source file starts with a standard two-line header:
    // Copyright <YEAR> <HOLDER>
    // SPDX-License-Identifier: <LICENSE>

USAGE
  go run ./tools/licenser/main.go [flags] <directories or files>

FLAGS
  -check-only           Only report non-compliant files; do not modify them (default: false)
  -c, --copyright <str> Copyright holder 
  -l, --license   <str> License identifier to write: "apache" (default)

EXAMPLES
  # Check license compliance in all Go files under the current directory
  go run ./tools/licenser/main.go -check-only -c "The OpenChoreo Authors" .

  # Add/fix license headers in place
  go run ./tools/licenser/main.go -c "The OpenChoreo Authors" .

LEARN MORE
  SPDX License Identifiers: https://spdx.org/licenses/

Note: This tool works with any source file type (e.g., .go, .js, .ts, .py, etc.)
`

func main() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usageText) }
	flag.Parse()

	if flag.NArg() == 0 || (*flagHolder == "" && !*flagCheckOnly) {
		flag.Usage()
		os.Exit(0)
	}

	mode := "CHECK"
	if !*flagCheckOnly {
		mode = "FIX"
	}
	fmt.Printf("Running in %s mode (apache license)\n", mode)

	var offending []string
	for _, dir := range flag.Args() {
		files, err := walk(dir, *flagHolder, !*flagCheckOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error scanning %s: %v\n", dir, err)
			os.Exit(2)
		}
		offending = append(offending, files...)
	}

	if *flagCheckOnly {
		if len(offending) > 0 {
			fmt.Println("❌ Missing or invalid headers:")
			for _, f := range offending {
				fmt.Printf(" • %s\n", f)
			}
			os.Exit(1)
		}
		fmt.Println("✅ All files have valid headers.")
	} else {
		if len(offending) > 0 {
			fmt.Println("🛠 Added headers to:")
			for _, f := range offending {
				fmt.Printf(" • %s\n", f)
			}
		} else {
			fmt.Println("✅ No changes needed – all headers already valid.")
		}
	}
}
