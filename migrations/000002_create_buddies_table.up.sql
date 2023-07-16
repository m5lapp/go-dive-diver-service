create table if not exists buddies (
    id            bigint primary key generated always as identity,
    version       integer not null default 1,
    created_at    timestamp(8) with time zone not null default now(),
    updated_at    timestamp(8) with time zone not null default now(),
    user_id       char(20) not null,
    buddy_user_id char(20) references divers(user_id) on delete set null,
    name          text     not null,
    email         citext,
    phone_number  text,
    organisation  text, 
    org_member_id text,
    notes         text,
    unique(user_id, buddy_user_id),
    unique(user_id, email)
);

create index if not exists buddies_user_id_idx
    on buddies using gin (to_tsvector('simple', user_id));
