create index concurrently if not exists comments_post_parent_created_id_idx
    on comments (post_id, parent_id, created_at desc, id desc);
