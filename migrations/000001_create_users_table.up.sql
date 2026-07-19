create table if not exists users (
    id         uuid primary key default gen_random_uuid(),
    name       text unique not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
