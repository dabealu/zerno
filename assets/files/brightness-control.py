#!/usr/bin/env python3
"""Control brightness of an external display via i2c-dev/ddcutil."""

import os
import os.path
import subprocess
import sys

VALUE_FILE = os.path.join(os.path.expanduser("~"), ".brightness.value")
DISPLAY = "1"
AT_MIN = "Already at min value"
AT_MAX = "Already at max value"


def run(cmd):
    return subprocess.run(cmd, capture_output=True, text=True)


def load_module():
    if "i2c_dev" not in run(["lsmod"]).stdout:
        print("loading module i2c-dev")
        r = run(["sudo", "modprobe", "i2c-dev"])
        if r.returncode != 0:
            sys.exit(r.stderr)


def get_value():
    try:
        with open(VALUE_FILE) as f:
            return int(f.read().strip())
    except (OSError, ValueError):
        pass
    r = run(["sudo", "ddcutil", "--display", DISPLAY, "getvcp", "10"])
    if r.returncode == 0:
        for line in r.stdout.splitlines():
            if "Brightness" in line:
                for word in line.split():
                    if word.isdigit():
                        return int(word)
    return 100


def save_value(val):
    try:
        with open(VALUE_FILE, "w") as f:
            f.write(str(val))
    except OSError as e:
        notify("Failed to save value")
        sys.exit(str(e))


def set_value(val):
    r = run(["sudo", "ddcutil", "--display", DISPLAY, "setvcp", "10", str(val)])
    if r.returncode != 0:
        notify("Failed to set value")
        sys.exit(r.stderr)
    notify(f"Set to {val}")


def notify(msg):
    run(["notify-send", "--app-name=brightness-control", "--urgency=low",
         "Brightness control", msg])


def clamp_step(val, increase):
    val += 33 if increase else -33
    if val >= 95 or val > 100:
        return 100
    if val <= 5 or val < 0:
        return 0
    return val


def main():
    if len(sys.argv) < 2:
        sys.exit("flag must be specified, one of: increase, decrease, min, max")
    load_module()
    action = sys.argv[1]
    if action == "increase":
        val = get_value()
        if val < 100:
            val = clamp_step(val, True)
            set_value(val)
        else:
            notify(AT_MAX)
        save_value(val)
    elif action == "decrease":
        val = get_value()
        if val > 0:
            val = clamp_step(val, False)
            set_value(val)
        else:
            notify(AT_MIN)
        save_value(val)
    elif action == "min":
        if get_value() > 0:
            set_value(0)
        else:
            notify(AT_MIN)
        save_value(0)
    elif action == "max":
        if get_value() < 100:
            set_value(100)
        else:
            notify(AT_MAX)
        save_value(100)
    else:
        sys.exit("flag must be specified, one of: increase, decrease, min, max")


if __name__ == "__main__":
    main()
