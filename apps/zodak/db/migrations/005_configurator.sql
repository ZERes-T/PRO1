CREATE TABLE IF NOT EXISTS quotes (
  id         BIGSERIAL PRIMARY KEY,
  user_id    BIGINT NOT NULL,
  kind       TEXT NOT NULL DEFAULT 'wardrobe',
  w_mm       INT  NOT NULL,
  h_mm       INT  NOT NULL,
  d_mm       INT  NOT NULL,
  material   TEXT NOT NULL,         -- ldsp | mdf | plywood
  facade     TEXT NOT NULL,         -- none | mirror | glass | mdf_paint
  drawers    INT  NOT NULL DEFAULT 0,
  shelves    INT  NOT NULL DEFAULT 0,
  doors      INT  NOT NULL DEFAULT 2,
  corner     BOOLEAN NOT NULL DEFAULT false,
  lighting   BOOLEAN NOT NULL DEFAULT false,
  total      NUMERIC(12,2) NOT NULL,
  breakdown  JSONB NOT NULL,
  status     TEXT NOT NULL DEFAULT 'draft',  -- draft | sent | accepted | archived
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_quotes_user ON quotes(user_id, created_at DESC);

