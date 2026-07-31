#!/usr/bin/env python3
"""Convert all HEIC files from the logbook2 source directory to JPG in logbook-2."""

import os
from pathlib import Path

import pillow_heif
from PIL import Image

pillow_heif.register_heif_opener()

SRC_DIR = Path(__file__).parent / "logbook2-20260327T154314Z-3-001" / "logbook2"
DST_DIR = Path(__file__).parent / "logbook-2"

DST_DIR.mkdir(exist_ok=True)

heic_files = sorted(SRC_DIR.glob("*.HEIC"))
print(f"Found {len(heic_files)} HEIC files in {SRC_DIR}")

for heic_path in heic_files:
    jpg_name = heic_path.stem + ".jpg"
    jpg_path = DST_DIR / jpg_name
    if jpg_path.exists():
        print(f"  Skipping {jpg_name} (already exists)")
        continue
    print(f"  Converting {heic_path.name} -> {jpg_name}")
    img = Image.open(heic_path)
    img.save(jpg_path, "JPEG", quality=95)

print("Done.")
