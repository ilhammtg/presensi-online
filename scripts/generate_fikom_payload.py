#!/usr/bin/env python3
import re
import json
import os

with open("dev-docs/contoh-jadwal.md", "r", encoding="utf-8") as f:
    lines = f.readlines()

current_class = None
raw_schedules = []

for line in lines:
    line = line.strip()
    m_class = re.match(r"^##\s+([A-Z0-9\-]+)", line)
    if m_class:
        current_class = m_class.group(1)
        continue
    if line.startswith("|") and not line.startswith("| Hari") and not line.startswith("|---"):
        parts = [p.strip() for p in line.split("|")[1:-1]]
        if len(parts) >= 4:
            hari, matkul, dosen, ruang = parts[0], parts[1], parts[2], parts[3]
            raw_schedules.append({
                "class": current_class,
                "hari": hari,
                "matkul": matkul,
                "dosen": dosen,
                "ruang": ruang if ruang != "–" else ("R1" if current_class.startswith("IM") else "R3")
            })

# 1. Unique Lecturers
unique_lecturers = sorted(list(set(s["dosen"] for s in raw_schedules)))
lecturer_map = {} # dosen_name -> {nidn, email, name, prodi}

for idx, name in enumerate(unique_lecturers, start=1):
    nidn = f"00{idx:02d}01850{idx%10}"
    # Generate clean email slug
    clean_name = re.sub(r"[^a-zA-Z\s]", "", name).strip().lower()
    name_parts = clean_name.split()
    first = name_parts[0] if name_parts else f"dosen{idx}"
    last = name_parts[-1] if len(name_parts) > 1 else "u"
    email = f"{first}.{last[:4]}@kampus.ac.id"
    
    # Determine prodi based on teaching classes
    classes_taught = [s["class"] for s in raw_schedules if s["dosen"] == name]
    is_im_only = all(c.startswith("IM") for c in classes_taught)
    prodi = "IM" if is_im_only else "INF"
    
    lecturer_map[name] = {
        "external_id": nidn,
        "name": name,
        "email": email,
        "role": "dosen",
        "prodi_code": prodi,
        "password": "password"
    }

# 2. Rooms
rooms_set = sorted(list(set(s["ruang"] for s in raw_schedules)))
room_data_list = []
for idx, r in enumerate(rooms_set):
    is_lab = r.startswith("L")
    lat = 5.194120 + (idx * 0.000010) if is_lab else 5.193730 + (idx * 0.000008)
    lon = 96.788100 + (idx * 0.000010) if is_lab else 96.787492 + (idx * 0.000008)
    name_desc = f"Laboratorium {r}" if is_lab else f"Ruang Kuliah {r}"
    if r == "RKU-1":
        name_desc = "Ruang Kuliah Umum 1"
    room_data_list.append({
        "room_code": r,
        "name": name_desc,
        "latitude": round(lat, 6),
        "longitude": round(lon, 6),
        "radius_meters": 80
    })

# 3. Classes and Students
classes_list = sorted(list(set(s["class"] for s in raw_schedules)))
students_by_class = {}
all_users = []

# Admins & Superadmin
all_users.append({
    "external_id": "SUPERADMIN01",
    "name": "Administrator Sistem",
    "email": "superadmin@kampus.ac.id",
    "role": "superadmin",
    "prodi_code": "INF",
    "password": "password"
})
all_users.append({
    "external_id": "ADM-001",
    "name": "Admin Prodi Informatika",
    "email": "admin.ti@kampus.ac.id",
    "role": "admin_prodi",
    "prodi_code": "INF",
    "password": "password"
})
all_users.append({
    "external_id": "ADM-IM-01",
    "name": "Admin Prodi Informatika Medis",
    "email": "admin.im@kampus.ac.id",
    "role": "admin_prodi",
    "prodi_code": "IM",
    "password": "password"
})

# Add all lecturers
for l in lecturer_map.values():
    all_users.append(l)

# Student name pools
first_names = ["Ahmad", "Muhammad", "Rizki", "Fajar", "Diki", "Cut", "Siti", "Nurul", "Putri", "Teuku", "Zul", "Intan", "Faisal", "Syahrul", "Maulana", "Rahmat", "Tari", "Nadia", "Dewi", "Bima"]
last_names = ["Pratama", "Hidayat", "Saputra", "Aulia", "Rahma", "Wulandari", "Maulida", "Munandar", "Akbar", "Fitriani", "Zulkarnain", "Suryani", "Wardani", "Iskandar", "Ramadhan"]

std_counter = 1
for c in classes_list:
    prodi_code = "IM" if c.startswith("IM") else "INF"
    sem_num = 1
    if "-3" in c: sem_num = 3
    elif "-5" in c: sem_num = 5
    elif "-7" in c: sem_num = 7
    year_prefix = 2026 - ((sem_num - 1) // 2)
    prodi_num = "55201" if prodi_code == "INF" else "55202"
    
    class_students = []
    # 5 students per class
    for s_idx in range(1, 6):
        nim = f"{year_prefix % 100}{prodi_num}{std_counter:04d}"
        fn = first_names[(std_counter + s_idx) % len(first_names)]
        ln = last_names[(std_counter * 2 + s_idx) % len(last_names)]
        full_name = f"{fn} {ln}"
        email = f"{nim}@mahasiswa.kampus.ac.id"
        
        all_users.append({
            "external_id": nim,
            "name": full_name,
            "email": email,
            "role": "mahasiswa",
            "prodi_code": prodi_code,
            "password": "password"
        })
        class_students.append(nim)
        std_counter += 1
    students_by_class[c] = class_students

# 4. Schedules
day_map = {
    "Senin": 1,
    "Selasa": 2,
    "Rabu": 3,
    "Kamis": 4,
    "Jumat": 5,
    "Sabtu": 6,
    "Senin/Selasa": 1,
    "Selasa/Rabu": 2,
    "Rabu/Kamis": 3,
    "Jumat/Sabtu": 5
}

time_slots = [
    ("08:00:00", "10:30:00"),
    ("10:45:00", "13:15:00"),
    ("14:00:00", "16:30:00"),
    ("16:45:00", "18:45:00"),
]

course_code_map = {}
course_counter = 101

# Track schedule per class & day to allocate non-overlapping time slots
class_day_slot_tracker = {}

schedules_payload = []
for idx, s in enumerate(raw_schedules, start=1):
    c_name = s["matkul"]
    if c_name not in course_code_map:
        prefix = "MED" if s["class"].startswith("IM") else "INF"
        course_code_map[c_name] = f"{prefix}{course_counter}"
        course_counter += 1
    
    c_code = course_code_map[c_name]
    lecturer_info = lecturer_map[s["dosen"]]
    day_num = day_map.get(s["hari"], 1)
    
    tracker_key = (s["class"], day_num)
    slot_idx = class_day_slot_tracker.get(tracker_key, 0)
    slot = time_slots[slot_idx % len(time_slots)]
    class_day_slot_tracker[tracker_key] = slot_idx + 1
    
    clean_cls = s["class"].replace("-", "")
    sched_id = f"SCH-2026-{clean_cls}-{c_code}-{idx}"
    
    schedules_payload.append({
        "external_schedule_id": sched_id,
        "course_code": c_code,
        "course_name": f"{c_name} ({s['class']})",
        "lecturer_id": lecturer_info["external_id"],
        "room_code": s["ruang"],
        "day_of_week": day_num,
        "start_time": slot[0],
        "end_time": slot[1],
        "enrolled_students": students_by_class[s["class"]]
    })

payload = {
    "sync_timestamp": "2026-09-25T00:00:00Z",
    "academic_year": "2026/2027",
    "semester_type": 1,
    "faculties": [
        {
            "code": "FIKOM",
            "name": "Fakultas Ilmu Komputer",
            "study_programs": [
                {"code": "INF", "name": "S1 Informatika"},
                {"code": "IM", "name": "S1 Informatika Medis"}
            ]
        }
    ],
    "buildings": [
        {
            "code": "GEDUNG-FIKOM",
            "name": "Gedung Fakultas Ilmu Komputer",
            "rooms": room_data_list
        }
    ],
    "users": all_users,
    "schedules": schedules_payload
}

out_path = "backend/cmd/mock-campus-api/payload.json"
os.makedirs(os.path.dirname(out_path), exist_ok=True)
with open(out_path, "w", encoding="utf-8") as f:
    json.dump(payload, f, indent=2, ensure_ascii=False)

print(f"✅ Generated payload successfully:")
print(f"   - File: {out_path}")
print(f"   - Total Faculties: {len(payload['faculties'])}")
print(f"   - Total Study Programs: {len(payload['faculties'][0]['study_programs'])}")
print(f"   - Total Rooms: {len(room_data_list)}")
print(f"   - Total Users: {len(all_users)} (Dosen: {len(lecturer_map)}, Mahasiswa: {len(all_users)-len(lecturer_map)-3}, Admins: 3)")
print(f"   - Total Schedules: {len(schedules_payload)}")
