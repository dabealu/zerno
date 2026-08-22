#!/usr/bin/env python3
"""nav - sway window icons for waybar.

Resident per-output daemon: subscribes to sway IPC, renders the focused
workspace's windows as pango icon spans (focused one chipped), emits one
JSON line per event on stdout for waybar's custom/nav module.
Pure stdlib. Single instance per $WAYBAR_OUTPUT_NAME.
"""

import json
import os
import signal
import socket
import struct
import time

MAGIC = b"i3-ipc"
SUBSCRIBE, GET_TREE = 2, 4
EVENT_BIT = 0x80000000

ACTIVE = "#ffffff"
CHIP_BG = "#3a3a3a"

BY_APP = {
    "google-chrome": "\uf0ac", "chromium": "\uf0ac", "firefox": "\uf0ac",
    "Alacritty": "\ue795", "alacritty": "\ue795", "kitty": "\ue795",
    "ghostty": "\ue795", "Ghostty": "\ue795",
    "thunar": "\uf07b", "Thunar": "\uf07b", "nautilus": "\uf07b",
    "org.telegram.desktop": "\uf1d8", "Telegram": "\uf1d8",
    "virt-manager": "\uf108",
    "Steam": "\uf1b7", "steam": "\uf1b7",
}


def glyph(app, title):
    if title.startswith("nvim"):
        return "\uf36f"
    if title.startswith("opencode") or title[:2].lower() == "oc":
        return "\uee15"
    return BY_APP.get(app, "\ueef6")


def span(glyph_, focused):
    bg = f' background="{CHIP_BG}"' if focused else ""
    return (f'<span foreground="{ACTIVE}"{bg} '
            f'font-size="16pt" rise="-2000"> {glyph_} </span>')


class Sway:
    """Minimal i3/sway IPC client: framed JSON messages over $SWAYSOCK."""

    def __init__(self):
        self.s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.s.connect(os.environ["SWAYSOCK"])

    def _read(self, n):
        buf = b""
        while len(buf) < n:
            chunk = self.s.recv(n - len(buf))
            if not chunk:
                raise ConnectionError("sway ipc closed")
            buf += chunk
        return buf

    def roundtrip(self, mtype, payload=b""):
        self.send(mtype, payload)
        while True:  # demux: events may race replies; replies carry bare type
            mtype_rx, data = self.recv()
            if not mtype_rx & EVENT_BIT:
                return data

    def send(self, mtype, payload=b""):
        self.s.sendall(MAGIC + struct.pack("=II", len(payload), mtype) + payload)

    def recv(self):
        head = self._read(14)
        length, mtype = struct.unpack("=II", head[6:14])
        return mtype, json.loads(self._read(length))


def leaves(node):
    kids = node.get("nodes", []) + node.get("floating_nodes", [])
    if not kids:
        yield node
        return
    for kid in kids:
        yield from leaves(kid)


def hasfocus(node):
    if node.get("focused"):
        return True
    return any(hasfocus(kid)
               for kid in node.get("nodes", []) + node.get("floating_nodes", []))


def render(sw):
    out_name = os.environ.get("WAYBAR_OUTPUT_NAME", "")
    tree = sw.roundtrip(GET_TREE)
    outputs = [o for o in tree.get("nodes", [])
               if o.get("name") == out_name
               or (not out_name and not o.get("name", "").startswith("__"))]
    if not outputs:
        print(json.dumps({"text": ""}), flush=True)
        return
    ws = next((w for w in outputs[0].get("nodes", []) if hasfocus(w)), None)
    parts = []
    if ws is not None:
        for leaf in leaves(ws):
            app = leaf.get("app_id") or \
                leaf.get("window_properties", {}).get("class", "")
            parts.append(span(glyph(app, leaf.get("name") or ""),
                              leaf.get("focused")))
    print(json.dumps({"text": "".join(parts)}), flush=True)


def kill_others():
    """Cull same-output instances only; other outputs' navs are legit."""
    mine = os.environ.get("WAYBAR_OUTPUT_NAME", "")
    me = os.getpid()
    for pid in os.listdir("/proc"):
        if not pid.isdigit() or int(pid) == me:
            continue
        try:
            with open(f"/proc/{pid}/cmdline", "rb") as fh:
                if "sway/nav.py" not in fh.read().decode(errors="ignore").replace("\0", " "):
                    continue
            with open(f"/proc/{pid}/environ", "rb") as fh:
                env = fh.read().decode(errors="ignore").split("\0")
            theirs = next((e.split("=", 1)[1] for e in env
                           if e.startswith("WAYBAR_OUTPUT_NAME=")), "")
        except OSError:
            continue
        if theirs == mine:
            try:
                os.kill(int(pid), signal.SIGKILL)
            except OSError:
                pass


def run():
    sw = Sway()
    res = sw.roundtrip(SUBSCRIBE, b'["window","workspace"]')
    if not (isinstance(res, dict) and res.get("success", True)):
        raise RuntimeError("subscribe rejected")
    render(sw)
    while True:
        mtype, _ = sw.recv()
        if mtype & EVENT_BIT:
            render(sw)


def main():
    kill_others()
    while True:
        try:
            run()
        except Exception:
            time.sleep(1)


if __name__ == "__main__":
    main()
