create schema if not exists public;

create table if not exists public.job_applications (
    id bigserial primary key,
    company_name varchar(255),
    position_title varchar(255) not null,
    source_url text,
    work_type varchar(50),
    salary varchar(100),
    position_description text,
    tech_stack text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists public.job_application_status_history (
    id bigserial primary key,
    application_id bigint not null references public.job_applications(id) on delete cascade,
    status varchar(50) not null,
    note text,
    changed_at timestamptz not null default now(),
    constraint job_application_status_history_status_check
        check (status in ('applied', 'screening', 'interview', 'offer', 'rejected', 'withdrawn', 'accepted'))
);

create index if not exists idx_job_application_status_history_application_id_changed_at
    on public.job_application_status_history (application_id, changed_at desc, id desc);

create table if not exists public.job_application_comments (
    id bigserial primary key,
    application_id bigint not null references public.job_applications(id) on delete cascade,
    body text not null,
    created_at timestamptz not null default now()
);

create index if not exists idx_job_application_comments_application_id_created_at
    on public.job_application_comments (application_id, created_at asc, id asc);

create table if not exists public.job_application_documents (
    id bigserial primary key,
    application_id bigint not null references public.job_applications(id) on delete cascade,
    document_type varchar(50) not null,
    original_filename varchar(255) not null,
    content_type varchar(100) not null,
    storage_path text not null unique,
    sha256 varchar(64) not null,
    size_bytes bigint not null check (size_bytes > 0),
    uploaded_at timestamptz not null default now(),
    constraint job_application_documents_document_type_check
        check (document_type in ('cv', 'cover_letter'))
);

create index if not exists idx_job_application_documents_application_id_uploaded_at
    on public.job_application_documents (application_id, uploaded_at desc, id desc);

alter table public.job_applications owner to openclaw;
alter table public.job_application_status_history owner to openclaw;
alter table public.job_application_comments owner to openclaw;
alter table public.job_application_documents owner to openclaw;
