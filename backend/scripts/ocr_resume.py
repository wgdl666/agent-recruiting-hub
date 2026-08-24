#!/usr/bin/env python3
"""Optional OCR fallback: render PDF page 1 and return placeholder text.
Install: pip install pymupdf (optional). Without it, returns empty."""
import sys

def main():
    if len(sys.argv) < 2:
        return
    path = sys.argv[1]
    try:
        import fitz  # pymupdf
        doc = fitz.open(path)
        text = ""
        for page in doc:
            text += page.get_text()
        if len(text.strip()) > 200:
            print(text)
            return
        # render first page for external OCR tools if needed
        pix = doc[0].get_pixmap(matrix=fitz.Matrix(2, 2))
        out = path + ".ocr.png"
        pix.save(out)
    except Exception:
        pass

if __name__ == "__main__":
    main()
