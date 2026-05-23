CREATE TABLE IF NOT EXISTS public.pastes (
    id VARCHAR(7) NOT NULL,
    code TEXT NOT NULL DEFAULT '',
    author_id BIGINT,
    CONSTRAINT pastes_pkey PRIMARY KEY (id)
);

ALTER TABLE public.pastes
    DROP CONSTRAINT IF EXISTS fkg9jxq5nlq789py3dehjha8w0n;

ALTER TABLE public.pastes
    DROP CONSTRAINT IF EXISTS fk_pastes_profile;

ALTER TABLE public.pastes
    ADD CONSTRAINT fk_pastes_profile
    FOREIGN KEY (author_id)
    REFERENCES public.profiles(id)
    ON DELETE SET NULL;

ALTER TABLE public.pastes
    ALTER COLUMN id TYPE CHAR(7) USING TRIM(id);

ANALYZE public.pastes;