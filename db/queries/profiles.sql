-- name: GetProfile :one
SELECT id FROM public.profiles
WHERE id = $1;

-- name: UpsertProfile :exec
INSERT INTO public.profiles (id)
VALUES ($1) ON CONFLICT DO NOTHING;
