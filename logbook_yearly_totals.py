import csv
from collections import defaultdict
from datetime import datetime

def parse_time_to_minutes(time_str):
    """Convert time string (H:MM) to total minutes."""
    if not time_str or time_str.strip() == '':
        return 0
    parts = str(time_str).strip().split(':')
    if len(parts) == 2:
        hours = int(parts[0])
        minutes = int(parts[1])
        return hours * 60 + minutes
    return 0

def minutes_to_hmm(total_minutes):
    """Convert total minutes back to H:MM format."""
    hours = total_minutes // 60
    minutes = total_minutes % 60
    return f"{hours}:{minutes:02d}"

def main():
    yearly_totals = defaultdict(lambda: {
        'total_time': 0,
        'pic_time': 0,
        'student_time': 0,
        'instrument_time': 0,
        'seaplane_time': 0,
        'landings': 0,
        'flights': 0
    })

    # Track previous cumulative seaplane time to calculate incremental time
    prev_seaplane_cumulative = None

    with open('logbook_1_final.csv', 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            # Parse date to extract year (format: DD/MM/YYYY)
            date_parts = row['Date'].split('/')
            year = int(date_parts[2])
            
            # Add times to yearly totals
            yearly_totals[year]['total_time'] += parse_time_to_minutes(row.get('Total_Time', '0'))
            yearly_totals[year]['pic_time'] += parse_time_to_minutes(row.get('PIC_Time', '0'))
            yearly_totals[year]['student_time'] += parse_time_to_minutes(row.get('Student_Time', '0'))
            yearly_totals[year]['instrument_time'] += parse_time_to_minutes(row.get('Instrument_Time', '0'))

            # Handle cumulative seaplane time - only add the difference from previous row
            current_seaplane_cumulative = parse_time_to_minutes(row.get('Cumulative_SEP_Sea', '0'))
            if prev_seaplane_cumulative is None:
                yearly_totals[year]['seaplane_time'] += current_seaplane_cumulative
            else:
                incremental = current_seaplane_cumulative - prev_seaplane_cumulative
                if incremental > 0:  # Only add if there was actual seaplane time this flight
                    yearly_totals[year]['seaplane_time'] += incremental
            prev_seaplane_cumulative = current_seaplane_cumulative

            yearly_totals[year]['landings'] += int(row.get('Landings', '0') or '0')
            yearly_totals[year]['flights'] += 1

    # Print results
    print(f"{'Year':<10} {'Flights':<10} {'Total Time':<15} {'PIC Time':<15} {'Student Time':<15} {'Instrument Time':<18} {'Seaplane Time':<18} {'Landings':<10}")
    print("-" * 120)

    for year in sorted(yearly_totals.keys()):
        data = yearly_totals[year]
        print(f"{year:<10} {data['flights']:<10} {minutes_to_hmm(data['total_time']):<15} {minutes_to_hmm(data['pic_time']):<15} {minutes_to_hmm(data['student_time']):<15} {minutes_to_hmm(data['instrument_time']):<18} {minutes_to_hmm(data['seaplane_time']):<18} {data['landings']:<10}")

    # Calculate grand totals
    grand_total = {
        'flights': sum(d['flights'] for d in yearly_totals.values()),
        'total_time': sum(d['total_time'] for d in yearly_totals.values()),
        'pic_time': sum(d['pic_time'] for d in yearly_totals.values()),
        'student_time': sum(d['student_time'] for d in yearly_totals.values()),
        'instrument_time': sum(d['instrument_time'] for d in yearly_totals.values()),
        'seaplane_time': sum(d['seaplane_time'] for d in yearly_totals.values()),
        'landings': sum(d['landings'] for d in yearly_totals.values())
    }
    
    print("-" * 120)
    print(f"{'GRAND TOTAL':<10} {grand_total['flights']:<10} {minutes_to_hmm(grand_total['total_time']):<15} {minutes_to_hmm(grand_total['pic_time']):<15} {minutes_to_hmm(grand_total['student_time']):<15} {minutes_to_hmm(grand_total['instrument_time']):<18} {minutes_to_hmm(grand_total['seaplane_time']):<18} {grand_total['landings']:<10}")

if __name__ == '__main__':
    main()