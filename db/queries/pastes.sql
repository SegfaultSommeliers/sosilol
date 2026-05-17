-- name: GetPaste :one
SELECT id, code, author_id FROM public.pastes
WHERE id = $1;

-- name: InsertPaste :exec
INSERT INTO public.pastes (id, code, author_id)
VALUES ($1, $2, $3);

-- name: GetPastesByAuthorID :many
SELECT id, code, author_id FROM public.pastes
WHERE author_id = $1;
