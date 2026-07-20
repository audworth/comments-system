create index concurrently if not exists posts_created_at_id_idx
    on posts (created_at desc, id desc);

