-- 002_task_config.sql
--
-- Dua perubahan konfigurasi tugas:
--   1. Kode dampak diseragamkan dengan prioritas menjadi HIGH, MEDIUM, LOW.
--      Makna lamanya tetap dicatat pada kolom label agar jejak iTop tidak
--      hilang: HIGH berarti dampak selingkup departemen, MEDIUM selingkup
--      layanan, dan LOW hanya perorangan.
--   2. Tabel prioritas mendapat kolom max_due_minutes, yaitu batas terlama
--      sebuah tugas boleh diberi tenggat, dihitung sejak tanggal mulai.
--
-- Skrip ini idempoten: aman dijalankan berulang kali. Jalankan setelah
-- 001_task_value_score.sql.
--   psql "<dsn>" -f db/migrations/002_task_config.sql

BEGIN;

-- 1. Batas tenggat per prioritas ------------------------------------------
-- Nilai 0 berarti tanpa batas, sehingga validasi tidak diberlakukan.
ALTER TABLE task_priorities ADD COLUMN IF NOT EXISTS max_due_minutes integer;

UPDATE task_priorities SET max_due_minutes = CASE upper(priority)
    WHEN 'HIGH'   THEN 1440   -- 1 hari
    WHEN 'MEDIUM' THEN 2880   -- 2 hari
    WHEN 'LOW'    THEN 4320   -- 3 hari
    ELSE 0
END
WHERE max_due_minutes IS NULL;

ALTER TABLE task_priorities ALTER COLUMN max_due_minutes SET DEFAULT 0;
UPDATE task_priorities SET max_due_minutes = 0 WHERE max_due_minutes IS NULL;
ALTER TABLE task_priorities ALTER COLUMN max_due_minutes SET NOT NULL;

-- 2. Penyeragaman kode dampak ---------------------------------------------
-- Kunci tamu memakai kolom no, bukan kode, sehingga penggantian kode ini
-- tidak memutus relasi mana pun.
UPDATE task_impacts SET impact = 'HIGH',   label = 'High'   WHERE no = 1;
UPDATE task_impacts SET impact = 'MEDIUM', label = 'Medium' WHERE no = 2;
UPDATE task_impacts SET impact = 'LOW',    label = 'Low'    WHERE no = 3;

-- Kolom kode pada tabel tugas mengikuti tabel acuannya.
UPDATE tasks t SET impact = ti.impact
FROM task_impacts ti
WHERE ti.no = t.impact_id AND t.impact IS DISTINCT FROM ti.impact;

ALTER TABLE tasks ALTER COLUMN impact SET DEFAULT 'HIGH';

COMMIT;

-- 3. Pemeriksaan hasil -----------------------------------------------------
--   SELECT no, priority, label, weight, max_due_minutes FROM task_priorities ORDER BY no;
--   SELECT no, impact, label, weight FROM task_impacts ORDER BY no;
--   SELECT impact, COUNT(*) FROM tasks GROUP BY impact ORDER BY 2 DESC;
--
-- Catatan: batas tenggat baru ini berlaku bagi tugas yang dibuat atau disunting
-- setelah migrasi. Tugas lama yang tenggatnya melampaui batas tidak diubah,
-- sehingga nilai OTR pada data penelitian tetap sebagaimana adanya.
