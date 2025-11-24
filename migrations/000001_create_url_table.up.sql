create table if not exists url (
    id uuid primary key default gen_random_uuid(),
    url_string text not null,
    url_code text not null unique,
    created_at timestamptz not null default now(),
    expire_at timestamptz not null default (now() + interval '1 day'),
    expired boolean not null default false
);

