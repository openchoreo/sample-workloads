// Copyright 2025 The OpenChoreo Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLicenseCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "License Check Suite")
}

var _ = Describe("License Header Checker", func() {
	var tmpDir string
	var header string

	const holder = "The OpenChoreo Authors"

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "license-check-test")
		Expect(err).NotTo(HaveOccurred())

		header = shortHeader(time.Now().Format("2006"), holder, "//")
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	writeFile := func(name, content string) string {
		p := filepath.Join(tmpDir, name)
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
		return p
	}

	It("detects a valid header", func() {
		content := header + "\n\npackage main\n\nfunc main() {}\n"
		path := writeFile("valid.go", content)

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("detects a missing header", func() {
		path := writeFile("missing.go", "package main\n\nfunc main() {}\n")

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("detects an incorrect holder", func() {
		bad := `// Copyright 2025 Someone Else
// SPDX-License-Identifier: Apache-2.0

package main

func main() {}
`
		path := writeFile("badholder.go", bad)

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("adds a header when missing", func() {
		path := writeFile("add.go", "package main\n\nfunc main() {}\n")

		commentPrefix := getCommentPrefix(path)
		header := shortHeader(time.Now().Format("2006"), holder, commentPrefix)
		updated, err := process(path, header, holder, true /* fix */)
		Expect(err).NotTo(HaveOccurred())
		Expect(updated).To(BeTrue())

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("reports non-compliance in check-only mode", func() {
		path := writeFile("checkonly.go", "package main\n\nfunc main() {}\n")

		commentPrefix := getCommentPrefix(path)
		header := shortHeader(time.Now().Format("2006"), holder, commentPrefix)
		updated, err := process(path, header, holder, false /* check only */)
		Expect(err).NotTo(HaveOccurred())
		Expect(updated).To(BeTrue()) // non-compliant

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeFalse())
	})

	It("walks a directory in check-only mode", func() {
		writeFile("walk1.go", "package main\n\nfunc main() {}\n")

		files, err := walk(tmpDir, holder, false /* check only */)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(strings.HasSuffix(files[0], "walk1.go")).To(BeTrue())
	})

	It("walks a directory and fixes headers", func() {
		writeFile("walk2.go", "package main\n\nfunc main() {}\n")

		files, err := walk(tmpDir, holder, true /* fix */)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(strings.HasSuffix(files[0], "walk2.go")).To(BeTrue())

		ok, err := hasValidHeader(files[0], holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("detects valid headers in Python files", func() {
		pythonHeader := shortHeader(time.Now().Format("2006"), holder, "#")
		content := pythonHeader + "\n\ndef main():\n    pass\n"
		path := writeFile("valid.py", content)

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("detects valid headers in JavaScript files", func() {
		jsHeader := shortHeader(time.Now().Format("2006"), holder, "//")
		content := jsHeader + "\n\nfunction main() {\n    console.log('hello');\n}\n"
		path := writeFile("valid.js", content)

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("adds headers to Python files", func() {
		path := writeFile("missing.py", "def main():\n    pass\n")

		commentPrefix := getCommentPrefix(path)
		header := shortHeader(time.Now().Format("2006"), holder, commentPrefix)
		updated, err := process(path, header, holder, true /* fix */)
		Expect(err).NotTo(HaveOccurred())
		Expect(updated).To(BeTrue())

		ok, err := hasValidHeader(path, holder)
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("walks directory and finds multiple file types", func() {
		writeFile("test.go", "package main\n\nfunc main() {}\n")
		writeFile("test.py", "def main():\n    pass\n")
		writeFile("test.js", "function main() {\n    console.log('hello');\n}\n")
		writeFile("ignored.txt", "This file should be ignored\n")

		files, err := walk(tmpDir, holder, false /* check only */)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(3)) // Only source files should be processed

		// Verify all source files were found
		filenames := make([]string, len(files))
		for i, f := range files {
			filenames[i] = filepath.Base(f)
		}
		Expect(filenames).To(ContainElements("test.go", "test.py", "test.js"))
	})
})
