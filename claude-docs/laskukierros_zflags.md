# laskukierros cross-check — `Z` flag mismatches (2026-08-01)

Generated from `laskukierros_export.csv`, whose times are **LOCAL** (proof in `drift.md`).
**Documented only — the CSVs were deliberately NOT changed** (user decision 2026-08-01; the paper
logbook stays authoritative). This list tells the future normalizing app which `Z` flags not to trust.

Matched 82 rows on date+registration+time: **52 are genuinely local**
(11 stored correctly, **41 wrongly carry `Z`**) and **30 are genuinely UTC**
(27 stored correctly, **3 are missing their `Z`**).

## A. Stored WITH `Z` but actually LOCAL (41 rows)
laskukierros time equals the paper time exactly — no UTC offset.

| file | date | reg | ours (stored) | laskukierros (local) |
|---|---|---|---|---|
| bk2 | 18/04/2021 | OH-COK | 07:25Z–08:38Z | 07:25–08:38 |
| bk2 | 18/04/2021 | OH-COK | 10:54Z–12:04Z | 10:54–12:04 |
| bk3 | 30/10/2021 | OH-CAM | 11:54Z–12:59Z | 11:54–12:59 |
| bk3 | 30/10/2021 | OH-CAM | 12:59Z–14:31Z | 12:59–14:31 |
| bk3 | 30/10/2021 | OH-CAM | 14:46Z–15:35Z | 14:46–15:35 |
| bk3 | 17/12/2021 | OH-CAM | 17:57Z–18:29Z | 17:57–18:27 |
| bk3 | 10/02/2022 | OH-CAM | 11:05Z–11:40Z | 11:05–11:40 |
| bk3 | 10/05/2022 | OH-CTL | 19:50Z–20:30Z | 19:50–20:30 |
| bk3 | 13/05/2022 | OH-CTL | 15:17Z–16:07Z | 15:17–16:07 |
| bk3 | 22/05/2022 | OH-CAM | 08:52Z–09:29Z | 08:52–09:29 |
| bk3 | 30/05/2022 | OH-CTL | 18:00Z–19:15Z | 18:00–19:15 |
| bk3 | 10/06/2022 | OH-CAM | 17:12Z–17:28Z | 17:12–17:28 |
| bk3 | 10/06/2022 | OH-CAM | 17:55Z–19:03Z | 17:55–19:03 |
| bk3 | 20/06/2022 | OH-CGX | 17:24Z–17:52Z | 17:24–17:52 |
| bk3 | 20/06/2022 | OH-CGX | 18:08Z–18:33Z | 18:08–18:33 |
| bk3 | 04/07/2022 | OH-CTL | 19:35Z–20:45Z | 19:35–20:45 |
| bk3 | 05/07/2022 | OH-TIL | 20:06Z–20:28Z | 20:06–20:28 |
| bk3 | 23/08/2022 | OH-CTL | 13:30Z–15:31Z | 13:30–15:35 |
| bk3 | 23/08/2022 | OH-CTL | 17:03Z–19:09Z | 17:03–19:09 |
| bk3 | 12/10/2022 | OH-CTL | 12:43Z–13:35Z | 12:43–13:35 |
| bk3 | 14/12/2022 | OH-CGX | 10:40Z–11:40Z | 10:40–11:40 |
| bk3 | 01/01/2023 | OH-CAM | 11:40Z–12:32Z | 11:40–12:32 |
| bk3 | 12/05/2023 | OH-CTL | 10:29Z–11:03Z | 10:29–11:03 |
| bk3 | 21/05/2023 | OH-CTL | 10:20Z–11:40Z | 10:20–11:40 |
| bk3 | 21/05/2023 | OH-CTL | 12:33Z–13:40Z | 12:33–13:40 |
| bk3 | 29/06/2023 | OH-CAY | 13:06Z–13:31Z | 13:06–13:31 |
| bk3 | 05/08/2023 | OH-CTL | 11:10Z–12:30Z | 11:10–12:30 |
| bk3 | 19/08/2023 | OH-CTL | 11:58Z–14:21Z | 11:58–14:21 |
| bk3 | 19/08/2023 | OH-CTL | 18:00Z–18:37Z | 18:00–18:37 |
| bk3 | 20/08/2023 | OH-CTL | 19:47Z–20:47Z | 19:47–20:47 |
| bk3 | 04/09/2023 | OH-CTL | 16:53Z–17:35Z | 16:53–17:35 |
| bk3 | 07/09/2023 | OH-CTL | 17:51Z–19:11Z | 17:51–19:11 |
| bk3 | 17/12/2023 | OH-AWB | 12:27Z–12:55Z | 12:27–12:55 |
| bk3 | 27/12/2023 | OH-CGX | 12:22Z–12:37Z | 12:22–12:37 |
| bk3 | 27/12/2023 | OH-CGX | 12:45Z–13:05Z | 12:45–13:05 |
| bk3 | 28/12/2023 | OH-CGX | 11:34Z–12:24Z | 11:34–12:24 |
| bk3 | 16/01/2024 | OH-AWB | 14:30Z–15:19Z | 14:30–15:19 |
| bk3 | 08/02/2024 | OH-AWB | 08:15Z–08:49Z | 08:15–08:49 |
| bk3 | 05/03/2024 | OH-CGX | 12:13Z–12:39Z | 12:13–12:39 |
| bk3 | 05/03/2024 | OH-CMU | 19:32Z–19:55Z | 19:32–19:55 |
| bk3 | 28/03/2024 | OH-CAM | 16:47Z–17:37Z | 16:47–17:37 |

## B. Stored WITHOUT `Z` but actually UTC (3 rows)
laskukierros time = paper time + the Finnish offset for that date (EET +2 / EEST +3).

| file | date | reg | ours (stored) | laskukierros (local) |
|---|---|---|---|---|
| bk2 | 08/04/2021 | OH-COK | 16:00–17:16 | 19:00–20:16 |
| bk3 | 25/06/2024 | OH-CTL | 12:18–12:58 | 15:18–15:58 |
| bk3 | 21/07/2024 | OH-CTL | 12:39–13:36 | 15:39–16:36 |
