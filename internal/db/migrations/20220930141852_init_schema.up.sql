CREATE SCHEMA IF NOT EXISTS public;

CREATE TABLE IF NOT EXISTS public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS IDX_LOGIN_USERS ON public.users (login);

CREATE TYPE status_order AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TABLE IF NOT EXISTS public.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number VARCHAR(255) UNIQUE NOT NULL,
    status status_order DEFAULT 'NEW',
    points NUMERIC(8, 2) DEFAULT 0,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
COMMENT ON COLUMN orders.status IS 'Statuses: NEW; PROCESSING; INVALID; PROCESSED';
CREATE INDEX IF NOT EXISTS IDX_NUMBER_ORDERS ON public.orders (number);
CREATE INDEX IF NOT EXISTS IDX_STATUS_ORDERS ON public.orders (status);
CREATE INDEX IF NOT EXISTS IDX_CREATEDAT ON public.orders (created_at);


CREATE TABLE IF NOT EXISTS public.score (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    total NUMERIC(8, 2) DEFAULT 0,
    user_id UUID UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    points NUMERIC(8, 2) NOT NULL,
    type SMALLINT NOT NULL,
    user_id UUID NOT NULL,
    order_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
COMMENT ON COLUMN transactions.type IS 'Type transaction: 1-increase; 2-decrease';
CREATE INDEX IF NOT EXISTS IDX_TYPE_TRANSACTIONS ON public.transactions (type);