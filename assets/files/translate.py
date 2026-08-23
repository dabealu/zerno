#!/usr/bin/env python3
"""Translate text (stdin or args) via Google Translate, notify the result."""

import json
import os
import subprocess
import sys
import urllib.parse
import urllib.request

URL = "https://translate.google.com/translate_a/single"


def notify(msg):
    subprocess.run(
        ["notify-send", "--app-name=translate", "--urgency=low", "Translation", msg]
    )


def read_input():
    if not os.isatty(0):
        return sys.stdin.read()
    if len(sys.argv) < 2:
        notify(f"missing input text, usage: {sys.argv[0]} input text")
        return ""
    return " ".join(sys.argv[1:])


def detect_lang(text):
    if text and ("a" <= text[0] <= "z" or "A" <= text[0] <= "Z"):
        return "en", "ru"
    return "ru", "en"


def translate(text):
    src, dst = detect_lang(text)
    query = urllib.parse.urlencode({"client": "at", "dt": "t", "dj": "1",
                                    "sl": src, "tl": dst, "q": text})
    req = urllib.request.Request(f"{URL}?{query}", method="POST", data=b"",
                                 headers={"Content-Type":
                                          "application/x-www-form-urlencoded;charset=utf-8"})
    with urllib.request.urlopen(req, timeout=5) as resp:
        payload = json.loads(resp.read())
    return " ".join(s.get("trans", "") for s in payload.get("sentences", []))


def main():
    text = read_input()
    if not text:
        return
    try:
        notify(translate(text))
    except Exception as e:
        notify(f"failed to translate: {e}")


if __name__ == "__main__":
    main()
