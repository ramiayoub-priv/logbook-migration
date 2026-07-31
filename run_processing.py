import subprocess
import json
import os
import csv
import glob
import re

OLLAMA_BIN = "/home/havoc/ollama-local/bin/ollama"
OLLAMA_ENV = {**os.environ, "OLLAMA_MODELS": "/home/havoc/models/ollama"}

IMAGE_FOLDER = "./logbook-2"
OUTPUT_FOLDER = "./logbook-2-csv"
PREVIOUS_CSV = "./logbook_1_final.csv"
MODEL = "qwen2.5vl:7b"

CSV_COLUMNS = [
    "Date", "Aircraft_Type", "Aircraft_Reg", "Departure", "Arrival",
    "Off_Block", "On_Block", "Takeoff", "Landing", "Block_Time", "Total_Time",
    "Instrument_Time", "Night_Time", "PIC_Time", "Student_Time",
    "Instructor_Time", "pic_name", "Landings", "Remarks",
    "Cumulative_Total", "Cumulative_PIC", "Cumulative_Student",
    "Cumulative_Instrument", "Cumulative_SEP_Sea", "Cumulative_Landings",
]

# Seaplane registrations that count toward SEP Sea cumulative
SEAPLANE_REGS = {"OH-CTL", "SE-GKT", "OH-GKT", "OH-PAX", "OH-MIL", "OH-CTE", "OH-CDK"}

PROMPT_TEMPLATE = """This is a Finnish EASA pilot logbook spread (left + right page).

IMPORTANT: Every logbook page has EXACTLY 8 flight rows. You MUST return exactly 8 objects in the JSON array. If you only see 7, look more carefully — one row may be faint or partially obscured.

Extract every flight row. Ignore page totals, subtotals, and signatures.

The LEFT page columns in order are:
1. Päiväys (Date) - format DD.MM.YYYY or DD/MM/YYYY
2. Ilma-aluksen tyyppi (Aircraft Type) - e.g. C152, C172, DA40, P28A, C206
3. Ilma-aluksen tunnus (Registration) - e.g. OH-KLS, OH-CTL
4. Lähtöpaikka (Departure) - ICAO code like EFHF, or a place name for seaplanes
5. Saapumispaikka (Arrival) - ICAO code or place name
6. Off-block aika (Off Block time) - HH:MM
7. On-block aika (On Block time) - HH:MM

The RIGHT page columns in order are:
8. Lentoaika (Total flight time) - H:MM
9. Mittariaika (Instrument time) - H:MM or blank
10. Yölentoaika (Night time) - H:MM or blank
11. Päällikkö (PIC/Commander time) - H:MM — this is a DURATION, not a name
12. Oppilas (Student/Dual time) - H:MM or blank
13. Opettaja (Instructor time) - H:MM or blank
14. Päällikön nimi (Name of PIC) - a person's NAME like "self" or an instructor name
15. Laskujen lukumäärä (Number of landings) - INTEGER, count the number written
16. Huomautuksia (Remarks) - any text/notes

Return a JSON array of EXACTLY 8 objects. Each flight is one object with these exact keys:
{{"Date","Aircraft_Type","Aircraft_Reg","Departure","Arrival","Off_Block","On_Block","Total_Time","Instrument_Time","Night_Time","PIC_Time","Student_Time","Instructor_Time","pic_name","Landings","Remarks"}}

Rules:
- Date format: DD/MM/YYYY
- Times in H:MM format
- Landings is an integer number (1, 2, 3, etc.), not zero unless truly zero
- If Päällikkö has a time and Oppilas is blank, then PIC_Time = that time, Student_Time = ""
- If Oppilas has a time and Päällikkö is blank, then Student_Time = that time, PIC_Time = ""
- pic_name: read from the name column. If flying as PIC, it says "self" or similar
- Use "" for blank fields. Do NOT use null.
- For seaplane rows the departure/arrival may be place names, not ICAO codes. That is OK, write what you see.
- There are ALWAYS 8 rows. Count them carefully.

Return ONLY the JSON array, nothing else."""


def parse_time_to_minutes(t):
    """Convert H:MM or HH:MM string to total minutes. Returns 0 for empty/invalid."""
    if not t or not isinstance(t, str):
        return 0
    t = t.strip()
    m = re.match(r'^(\d+):(\d{2})$', t)
    if m:
        return int(m.group(1)) * 60 + int(m.group(2))
    return 0


def minutes_to_time(mins):
    """Convert total minutes to H:MM string."""
    if mins == 0:
        return ""
    h = mins // 60
    m = mins % 60
    return f"{h}:{m:02d}"


def get_last_cumulative(csv_path):
    """Read the last row of a CSV to get cumulative totals."""
    with open(csv_path, "r") as f:
        reader = csv.DictReader(f)
        last = None
        for row in reader:
            last = row
    if last is None:
        return {
            "Cumulative_Total": 0, "Cumulative_PIC": 0, "Cumulative_Student": 0,
            "Cumulative_Instrument": 0, "Cumulative_SEP_Sea": 0, "Cumulative_Landings": 0,
        }
    return {
        "Cumulative_Total": parse_time_to_minutes(last.get("Cumulative_Total", "")),
        "Cumulative_PIC": parse_time_to_minutes(last.get("Cumulative_PIC", "")),
        "Cumulative_Student": parse_time_to_minutes(last.get("Cumulative_Student", "")),
        "Cumulative_Instrument": parse_time_to_minutes(last.get("Cumulative_Instrument", "")),
        "Cumulative_SEP_Sea": parse_time_to_minutes(last.get("Cumulative_SEP_Sea", "")),
        "Cumulative_Landings": int(last.get("Cumulative_Landings", "0") or "0"),
    }


def run_ollama(image_path, prompt):
    cmd = [OLLAMA_BIN, "run", MODEL, prompt, image_path]
    result = subprocess.run(cmd, capture_output=True, text=True, env=OLLAMA_ENV)
    return result.stdout


def extract_json(text):
    try:
        start = text.index("[")
        end = text.rindex("]") + 1
        return json.loads(text[start:end])
    except (ValueError, json.JSONDecodeError):
        return None


def compute_cumulative(rows, cum):
    """Add cumulative columns to each row, updating cum dict in place."""
    for row in rows:
        total_mins = parse_time_to_minutes(row.get("Total_Time", ""))
        pic_mins = parse_time_to_minutes(row.get("PIC_Time", ""))
        student_mins = parse_time_to_minutes(row.get("Student_Time", ""))
        instr_mins = parse_time_to_minutes(row.get("Instrument_Time", ""))
        landings = int(row.get("Landings", 0) or 0)

        # If neither PIC nor Student time is set, default to PIC
        if pic_mins == 0 and student_mins == 0 and total_mins > 0:
            pic_mins = total_mins
            row["PIC_Time"] = row.get("Total_Time", "")
            if not row.get("pic_name"):
                row["pic_name"] = "self"

        cum["Cumulative_Total"] += total_mins
        cum["Cumulative_PIC"] += pic_mins
        cum["Cumulative_Student"] += student_mins
        cum["Cumulative_Instrument"] += instr_mins
        cum["Cumulative_Landings"] += landings

        # SEP Sea: add total_time if registration is a known seaplane
        reg = (row.get("Aircraft_Reg") or "").strip().upper()
        if reg in SEAPLANE_REGS:
            cum["Cumulative_SEP_Sea"] += total_mins

        row["Cumulative_Total"] = minutes_to_time(cum["Cumulative_Total"])
        row["Cumulative_PIC"] = minutes_to_time(cum["Cumulative_PIC"])
        row["Cumulative_Student"] = minutes_to_time(cum["Cumulative_Student"])
        row["Cumulative_Instrument"] = minutes_to_time(cum["Cumulative_Instrument"])
        row["Cumulative_SEP_Sea"] = minutes_to_time(cum["Cumulative_SEP_Sea"])
        row["Cumulative_Landings"] = str(cum["Cumulative_Landings"])

        # Fill in Block_Time = Total_Time if not set, and blank Takeoff/Landing
        if not row.get("Block_Time"):
            row["Block_Time"] = row.get("Total_Time", "")
        row.setdefault("Takeoff", "")
        row.setdefault("Landing", "")

    return rows


def save_csv(rows, output_path):
    with open(output_path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=CSV_COLUMNS, quoting=csv.QUOTE_ALL)
        writer.writeheader()
        for row in rows:
            clean = {col: row.get(col, "") for col in CSV_COLUMNS}
            writer.writerow(clean)


def main():
    os.makedirs(OUTPUT_FOLDER, exist_ok=True)

    # Get sorted image list (numeric order by IMG number)
    images = sorted(
        glob.glob(os.path.join(IMAGE_FOLDER, "*.jpg")),
        key=lambda p: int(os.path.basename(p).split("_")[1].split(".")[0]),
    )
    print(f"Found {len(images)} images in {IMAGE_FOLDER}")

    # Start with cumulative totals from previous logbook
    cum = get_last_cumulative(PREVIOUS_CSV)
    print(f"Starting cumulative totals from {PREVIOUS_CSV}:")
    print(f"  Total={minutes_to_time(cum['Cumulative_Total'])} "
          f"PIC={minutes_to_time(cum['Cumulative_PIC'])} "
          f"Student={minutes_to_time(cum['Cumulative_Student'])} "
          f"Instrument={minutes_to_time(cum['Cumulative_Instrument'])} "
          f"SEP_Sea={minutes_to_time(cum['Cumulative_SEP_Sea'])} "
          f"Landings={cum['Cumulative_Landings']}")

    failed = []
    total_rows = 0

    for image_path in images:
        basename = os.path.splitext(os.path.basename(image_path))[0]
        output_csv = os.path.join(OUTPUT_FOLDER, f"logbook_{basename}.csv")

        print(f"\nProcessing {os.path.basename(image_path)} ...")

        output = run_ollama(image_path, PROMPT_TEMPLATE)
        data = extract_json(output)

        if data:
            # Python computes all cumulative totals
            data = compute_cumulative(data, cum)

            save_csv(data, output_csv)
            print(f"  -> {len(data)} rows written to {output_csv}")
            total_rows += len(data)
        else:
            print(f"  ** FAILED to extract JSON")
            print(f"  Raw output: {output[:500]}")
            failed.append(os.path.basename(image_path))

    print(f"\n{'='*60}")
    print(f"Done. Total rows extracted: {total_rows}")
    if failed:
        print(f"Failed images: {failed}")
    print(f"Final cumulative: Total={minutes_to_time(cum['Cumulative_Total'])} "
          f"PIC={minutes_to_time(cum['Cumulative_PIC'])} "
          f"Landings={cum['Cumulative_Landings']}")


if __name__ == "__main__":
    main()