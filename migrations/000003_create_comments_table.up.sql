create table if not exists comments (
    id         uuid primary key default gen_random_uuid(),
    post_id    uuid not null references posts(id) on delete restrict,
    parent_id  uuid,
    author_id  uuid not null references users(id) on delete restrict,
    body       varchar(2000) not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    unique (post_id, id),

    foreign key (post_id, parent_id)
        references comments(post_id, id)
        on delete restrict,

    check (PARENT_ID is null or parent_id <> id)
);
