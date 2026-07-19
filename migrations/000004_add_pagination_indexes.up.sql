create index if not exists concurrently posts_created_at_id_idx
    on posts (created_at desc, id desc);

-- create index if not exists concurrently comments_post_parent_created_id_idx
--     on comments (post_id, parent_id, created_at desc, id desc);
