#!/usr/bin/env python3
"""抽简历文字：优先 PDF 文本层，不够再 OCR。stdout 打印正文。"""
import os
import subprocess
import sys


def pdftotext(path: str) -> str:
    try:
        out = subprocess.check_output(
            ["pdftotext", "-layout", "-enc", "UTF-8", path, "-"],
            stderr=subprocess.DEVNULL,
        )
        return out.decode("utf-8", "ignore")
    except Exception:
        return ""


def pymupdf_text(path: str) -> str:
    try:
        import fitz
    except Exception:
        return ""
    try:
        doc = fitz.open(path)
        return "".join(page.get_text() for page in doc)
    except Exception:
        return ""


def tesseract_pages(path: str) -> str:
    try:
        import fitz
    except Exception:
        return ""
    try:
        doc = fitz.open(path)
        parts = []
        for i, page in enumerate(doc):
            if i >= 3:
                break
            pix = page.get_pixmap(matrix=fitz.Matrix(2, 2))
            png = path + f".ocr-{i}.png"
            pix.save(png)
            try:
                out = subprocess.check_output(
                    ["tesseract", png, "stdout", "-l", "chi_sim+eng"],
                    stderr=subprocess.DEVNULL,
                )
                parts.append(out.decode("utf-8", "ignore"))
            except Exception:
                pass
            finally:
                try:
                    os.remove(png)
                except OSError:
                    pass
        return "\n".join(parts)
    except Exception:
        return ""


def main() -> None:
    if len(sys.argv) < 2:
        return
    path = sys.argv[1]
    text = pymupdf_text(path)
    if len(text.strip()) < 300:
        alt = pdftotext(path)
        if len(alt.strip()) > len(text.strip()):
            text = alt
    if len(text.strip()) < 300:
        ocr = tesseract_pages(path)
        if len(ocr.strip()) > len(text.strip()):
            text = ocr
    sys.stdout.write(text)


if __name__ == "__main__":
    main()
