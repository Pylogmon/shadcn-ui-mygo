#!/usr/bin/env python3
"""Regenerate only the source-supported Space Grotesk 600 native face.

The pinned original has no 600 named instance and its author publishes static
300/400/500/700 only. MyGo's macOS matcher therefore selected the same face at
500 and 600 in the native workflow. Other original variable files stay intact.
This optional restoration tool needs fonttools==4.60.1; go generate itself
only verifies the committed instance and needs no Python package.
"""

import hashlib
import json
from pathlib import Path

import fontTools
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont

if fontTools.__version__ != "4.60.1":
    raise SystemExit("Use the fixed generator: fonttools==4.60.1")

root = Path(__file__).resolve().parents[2]
manifest = json.loads((root / "manifest.json").read_text())
family = next(f for f in manifest["families"] if f["id"] == "space-grotesk")
source = next(s for s in family["sources"] if s["asset"].endswith("normal.ttf"))
original = root / source["asset"]
if hashlib.sha256(original.read_bytes()).hexdigest() != source["sha256"]:
    raise SystemExit("The fixed original Space Grotesk font changed")

font = TTFont(original, recalcTimestamp=False)
coordinates = {axis.axisTag: axis.defaultValue for axis in font["fvar"].axes}
coordinates["wght"] = 600
instantiateVariableFont(font, coordinates, inplace=True)
if "fvar" in font:
    raise SystemExit("The generated face must be fully static")

# Keep the original copyright and license records. Give the generated face a
# distinct internal name; public registration joins it to its source family.
for name_id in (1, 2, 3, 4, 6, 16, 17, 18, 21, 22, 25):
    font["name"].removeNames(nameID=name_id)
names = {
    1: "Native Go Font 13",
    2: "SemiBold",
    3: "NativeGoFont13-SemiBold-v2.000",
    4: "Native Go Font 13 SemiBold",
    6: "NativeGoFont13-SemiBold",
    16: "Native Go Font 13",
    17: "SemiBold",
    18: "Native Go Font 13 SemiBold",
    21: "Native Go Font 13",
    22: "SemiBold",
}
for name_id, text in names.items():
    font["name"].setName(text, name_id, 3, 1, 1033)
    font["name"].setName(text, name_id, 1, 0, 0)
font["OS/2"].usWeightClass = 600
font["OS/2"].fsSelection &= ~(1 | 32 | 64)
font["head"].macStyle &= ~3
destination = root / "assets/space-grotesk/semibold-native.ttf"
font.save(destination)
print(destination.name, hashlib.sha256(destination.read_bytes()).hexdigest())
