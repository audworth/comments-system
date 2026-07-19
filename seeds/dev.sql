begin;

truncate table comments, posts, users;

-- 10 пользователей
insert into users (id, name, created_at, updated_at)
select
    ('00000000-0000-0000-0000-' || lpad(user_number::text, 12, '0'))::uuid,
    format('user_%s', user_number),
    '2026-01-01 00:00:00+00'::timestamptz + make_interval(secs => user_number),
    '2026-01-01 00:00:00+00'::timestamptz + make_interval(secs => user_number)
from generate_series(1, 10) as user_number;

-- 10 горячих постов
insert into posts (
    id,
    author_id,
    title,
    body,
    comments_enabled,
    created_at,
    updated_at
)
select
    ('10000000-0000-0000-0000-' || lpad(post_number::text, 12, '0'))::uuid,
    ('00000000-0000-0000-0000-' || lpad(post_number::text, 12, '0'))::uuid,
    format('hot_post_%s', post_number),
    format('hot_post_body_%s', post_number),
    true,
    '2026-02-01 00:00:00+00'::timestamptz + make_interval(days => post_number),
    '2026-02-01 00:00:00+00'::timestamptz + make_interval(days => post_number)
from generate_series(1, 10) as post_number;

-- 1000 комментариев на каждый пост
insert into comments (
    id,
    post_id,
    parent_id,
    author_id,
    body,
    created_at,
    updated_at
)
select
    (
        '20000000-' || lpad(post_number::text, 4, '0') ||
        '-0000-0000-' || lpad(comment_number::text, 12, '0')
    )::uuid,
    ('10000000-0000-0000-0000-' || lpad(post_number::text, 12, '0'))::uuid,
    null,
    (
        '00000000-0000-0000-0000-' ||
        lpad((((post_number + comment_number - 2) % 10) + 1)::text, 12, '0')
    )::uuid,
    format('comment_%s_%s', post_number, comment_number),
    '2026-03-01 00:00:00+00'::timestamptz
        + make_interval(days => post_number, secs => comment_number),
    '2026-03-01 00:00:00+00'::timestamptz
        + make_interval(days => post_number, secs => comment_number)
from generate_series(1, 10) as post_number
cross join generate_series(1, 1000) as comment_number;

-- 100 ответов на каждый корневой комментарий
insert into comments (
    id,
    post_id,
    parent_id,
    author_id,
    body,
    created_at,
    updated_at
)
select
    (
        '30000000-' || lpad(post_number::text, 4, '0') ||
        '-0000-0000-' || lpad((comment_number * 100 + reply_number)::text, 12, '0')
    )::uuid,
    ('10000000-0000-0000-0000-' || lpad(post_number::text, 12, '0'))::uuid,
    (
        '20000000-' || lpad(post_number::text, 4, '0') ||
        '-0000-0000-' || lpad(comment_number::text, 12, '0')
    )::uuid,
    (
        '00000000-0000-0000-0000-' ||
        lpad((((post_number + comment_number + reply_number - 3) % 10) + 1)::text, 12, '0')
    )::uuid,
    format('reply_%s_%s_%s', post_number, comment_number, reply_number),
    '2026-04-01 00:00:00+00'::timestamptz
        + make_interval(days => post_number, secs => comment_number)
        + reply_number * interval '1 microsecond',
    '2026-04-01 00:00:00+00'::timestamptz
        + make_interval(days => post_number, secs => comment_number)
        + reply_number * interval '1 microsecond'
from generate_series(1, 10) as post_number
cross join generate_series(1, 1000) as comment_number
cross join generate_series(1, 100) as reply_number;

-- ветка глубиной 100
insert into comments (
    id,
    post_id,
    parent_id,
    author_id,
    body,
    created_at,
    updated_at
)
values (
    '40000000-0001-0000-0000-000000000000'::uuid,
    '10000000-0000-0000-0000-000000000001'::uuid,
    null,
    '00000000-0000-0000-0000-000000000001'::uuid,
    'deep_thread_root',
    '2026-05-01 00:00:00+00'::timestamptz,
    '2026-05-01 00:00:00+00'::timestamptz
);

insert into comments (
    id,
    post_id,
    parent_id,
    author_id,
    body,
    created_at,
    updated_at
)
select
    (
        '40000000-0001-0000-0000-' ||
        lpad(depth::text, 12, '0')
    )::uuid,
    '10000000-0000-0000-0000-000000000001'::uuid,
    (
        '40000000-0001-0000-0000-' ||
        lpad((depth - 1)::text, 12, '0')
    )::uuid,
    (
        '00000000-0000-0000-0000-' ||
        lpad((((depth - 1) % 10) + 1)::text, 12, '0')
    )::uuid,
    format('deep_reply_%s', depth),
    '2026-05-01 00:00:00+00'::timestamptz
        + depth * interval '1 microsecond',
    '2026-05-01 00:00:00+00'::timestamptz
        + depth * interval '1 microsecond'
from generate_series(1, 100) as depth
order by depth;

commit;
