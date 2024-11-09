CREATE TABLE IF NOT EXISTS public.mint_tx (
	tx_hash text NOT NULL,
    block_num integer NOT NULL,
	block_hash text NOT NULL,
	sender text NOT NULL,
	CONSTRAINT mint_tx_pk PRIMARY KEY (tx_hash)
);