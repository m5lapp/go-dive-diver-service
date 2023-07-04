create table if not exists divers (
    id                      bigserial primary key,
    version                 integer not null default 1,
    email                   citext unique not null,
    diving_since            date,
    dive_number_offset      smallint not null default 0, 
    default_diving_country  text,
    default_diving_timezone text
)