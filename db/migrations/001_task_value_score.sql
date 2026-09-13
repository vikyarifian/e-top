-- 001_task_value_score.sql
--
-- Menyiapkan skema untuk indikator Task Value Score (TVS), pengganti
-- Task Priority Score. Nilai sebuah tugas dihitung sebagai hasil kali
-- bobot prioritas dan bobot dampak, sehingga berkisar 0,640 sampai 1,000.
--
-- Skrip ini idempoten: aman dijalankan berulang kali.
-- Pemakaian:  psql "<dsn>" -f db/migrations/001_task_value_score.sql

BEGIN;

-- 1. Bobot prioritas -------------------------------------------------------
ALTER TABLE task_priorities ADD COLUMN IF NOT EXISTS weight numeric(4,3);

UPDATE task_priorities SET weight = CASE upper(priority)
    WHEN 'HIGH'   THEN 1.000
    WHEN 'MEDIUM' THEN 0.900
    WHEN 'LOW'    THEN 0.800
    ELSE 0.800
END;

ALTER TABLE task_priorities ALTER COLUMN weight SET NOT NULL;

-- 2. Tabel acuan dampak ----------------------------------------------------
-- Dimensi ini diadopsi dari kolom impact milik iTop: seberapa luas pihak
-- yang terkena bila tugas tidak tuntas.
CREATE TABLE IF NOT EXISTS task_impacts (
    no     integer PRIMARY KEY,
    impact text        NOT NULL UNIQUE,
    label  text        NOT NULL,
    color  text        NOT NULL DEFAULT '',
    value  bigint      NOT NULL DEFAULT 0,
    level  bigint      NOT NULL DEFAULT 0,
    weight numeric(4,3) NOT NULL
);

INSERT INTO task_impacts (no, impact, label, color, value, level, weight) VALUES
    (1, 'DEPARTMENT', 'Departemen', 'red',    1, 3, 1.000),
    (2, 'SERVICE',    'Layanan',    'yellow', 2, 2, 0.900),
    (3, 'PERSON',     'Perorangan', 'green',  3, 1, 0.800)
ON CONFLICT (no) DO UPDATE
    SET impact = EXCLUDED.impact,
        label  = EXCLUDED.label,
        color  = EXCLUDED.color,
        value  = EXCLUDED.value,
        level  = EXCLUDED.level,
        weight = EXCLUDED.weight;

-- 3. Relasi dampak pada tugas ---------------------------------------------
-- Kolom impact sebelumnya menyimpan angka 1..3 warisan iTop. Angka itu
-- dipindahkan ke impact_id, lalu kolom impact diubah menjadi kode teks agar
-- terbaca tanpa perlu menghafal arti angkanya.
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS impact_id integer;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'tasks' AND column_name = 'impact'
                 AND data_type IN ('smallint', 'integer', 'bigint')) THEN
        EXECUTE 'UPDATE tasks SET impact_id = COALESCE(NULLIF(impact, 0), 1) WHERE impact_id IS NULL';
        EXECUTE 'ALTER TABLE tasks ALTER COLUMN impact DROP DEFAULT';
        EXECUTE 'ALTER TABLE tasks ALTER COLUMN impact TYPE text USING
                    CASE impact WHEN 2 THEN ''SERVICE'' WHEN 3 THEN ''PERSON'' ELSE ''DEPARTMENT'' END';
    ELSIF NOT EXISTS (SELECT 1 FROM information_schema.columns
                      WHERE table_name = 'tasks' AND column_name = 'impact') THEN
        EXECUTE 'ALTER TABLE tasks ADD COLUMN impact text';
    END IF;
END $$;

-- Tugas tanpa nilai dampak diperlakukan sebagai dampak terluas agar tidak
-- diuntungkan oleh ketiadaan data.
UPDATE tasks SET impact_id = 1 WHERE impact_id IS NULL;
UPDATE tasks SET impact = 'DEPARTMENT' WHERE impact IS NULL OR impact = '';

ALTER TABLE tasks ALTER COLUMN impact_id SET DEFAULT 1;
ALTER TABLE tasks ALTER COLUMN impact SET DEFAULT 'DEPARTMENT';

-- Kunci tamu ditambahkan hanya bila belum ada.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tasks_task_impacts') THEN
        ALTER TABLE tasks ADD CONSTRAINT fk_tasks_task_impacts
            FOREIGN KEY (impact_id) REFERENCES task_impacts(no);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_tasks_impact_id ON tasks (impact_id);

COMMIT;

-- 4. Pemeriksaan hasil -----------------------------------------------------
-- Kedua kueri berikut harus mengembalikan 0 baris yatim dan sebaran bobot
-- yang wajar.
--   SELECT COUNT(*) FROM tasks t LEFT JOIN task_impacts ti ON ti.no = t.impact_id WHERE ti.no IS NULL;
--   SELECT tp.priority, ti.impact, tp.weight * ti.weight AS nilai_tugas, COUNT(*)
--     FROM tasks t JOIN task_priorities tp ON tp.no = t.priority_id
--                  JOIN task_impacts   ti ON ti.no = t.impact_id
--    GROUP BY 1,2,3 ORDER BY 3 DESC;
