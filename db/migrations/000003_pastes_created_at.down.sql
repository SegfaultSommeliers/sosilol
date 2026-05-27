DROP INDEX IF EXISTS idx_pastes_author_id_created_at;

ALTER TABLE public.pastes
    DROP COLUMN IF EXISTS created_at;
