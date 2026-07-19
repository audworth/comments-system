create table if not exists posts (
    id               uuid primary key default gen_random_uuid(),
    author_id        uuid not null references users(id) on delete restrict,
    title            text not null,
    body             text not null,
    comments_enabled boolean not null default true,
    created_at       timestamptz not null default now(),
    updated_at       timestamptz not null default now()
);
