create table if not exists divers (
    user_id                 char(20) primary key,
    version                 integer not null default 1,
    diving_since            date,
    dive_number_offset      smallint not null default 0, 
    default_diving_country  text,
    default_diving_timezone text
);

create index if not exists divers_user_id_idx
    on divers using gin (to_tsvector('simple', user_id));
