ALTER TABLE public.pastes
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX IF NOT EXISTS idx_pastes_author_id_created_at
    ON public.pastes (author_id, created_at DESC);
