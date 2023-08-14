create table if not exists agencies (
    id          bigint primary key generated always as identity,
    common_name text not null unique,
    full_name   text not null unique,
    acronym     text,
    url         text,
);

create table if not exists agency_courses (
    id                  bigint  primary key generated always as identity,
    agency_id           bigint,
    name                text    not null,
    url                 text,
    is_specialty_course boolean not null,
    is_tech_course      boolean not null,
    is_pro_course       boolean not null,
    unique(agency_id, name),
    foreign key (agency_id) references agencies(id) on delete cascade
);
