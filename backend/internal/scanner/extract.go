package scanner

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
)

var projectRoot string

// SetRoot 让抽字脚本走仓库/部署目录里的 scripts/，不依赖启动时的 cwd。
func SetRoot(root string) {
	projectRoot = strings.TrimSpace(root)
}

func textEnough(s string) bool {
	return utf8.RuneCountInString(strings.TrimSpace(s)) >= 300
}

func betterText(cur, next string) string {
	if utf8.RuneCountInString(strings.TrimSpace(next)) > utf8.RuneCountInString(strings.TrimSpace(cur)) {
		return next
	}
	return cur
}

func ExtractText(path string) (string, error) {
	best, err := extractPDFText(path)
	if err != nil {
		best = ""
	}
	best = betterText(best, pdftotext(path))
	if textEnough(best) {
		return best, nil
	}
	if ocr, ocrErr := tryOCR(path); ocrErr == nil {
		best = betterText(best, ocr)
	} else {
		log.Printf("scanner: ocr fallback skipped: %v", ocrErr)
	}
	if textEnough(best) {
		return best, nil
	}
	return best, nil
}

func pdftotext(path string) string {
	cmd := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", path, "-")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func tryOCR(path string) (string, error) {
	candidates := []string{"scripts/ocr_resume.py"}
	if projectRoot != "" {
		candidates = []string{
			filepath.Join(projectRoot, "scripts", "ocr_resume.py"),
			filepath.Join(projectRoot, "backend", "scripts", "ocr_resume.py"),
		}
	}
	var script string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			script = p
			break
		}
	}
	if script == "" {
		return "", os.ErrNotExist
	}
	cmd := exec.Command("python3", script, path)
	if projectRoot != "" {
		cmd.Dir = projectRoot
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func renderPDFPages(path string, maxPages int) [][]byte {
	if maxPages <= 0 {
		maxPages = 3
	}
	dir, err := os.MkdirTemp("", "resume-pages-*")
	if err != nil {
		return nil
	}
	defer os.RemoveAll(dir)
	prefix := filepath.Join(dir, "p")
	cmd := exec.Command("pdftoppm", "-png", "-r", "140", "-f", "1", "-l", fmt.Sprintf("%d", maxPages), path, prefix)
	if err := cmd.Run(); err != nil {
		return nil
	}
	var pages [][]byte
	for i := 1; i <= maxPages; i++ {
		candidates := []string{
			fmt.Sprintf("%s-%d.png", prefix, i),
			fmt.Sprintf("%s-%02d.png", prefix, i),
		}
		for _, p := range candidates {
			b, err := os.ReadFile(p)
			if err == nil && len(b) > 0 {
				pages = append(pages, b)
				break
			}
		}
	}
	return pages
}

func extractPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	total := r.NumPage()
	for i := 1; i <= total; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteByte('\n')
	}
	return buf.String(), nil
}
