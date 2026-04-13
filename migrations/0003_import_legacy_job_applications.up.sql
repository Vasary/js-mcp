create temporary table tmp_legacy_job_applications_import (
    id bigint primary key,
    company_name varchar(255),
    position_title varchar(255) not null,
    source_url text,
    work_type varchar(50),
    salary varchar(100),
    position_description text,
    tech_stack text,
    normalized_status varchar(50) not null,
    status_note text,
    created_ts timestamptz not null,
    updated_ts timestamptz not null,
    import_note text,
    import_note_created_at timestamptz not null
) on commit drop;

insert into tmp_legacy_job_applications_import (
    id,
    company_name,
    position_title,
    source_url,
    work_type,
    salary,
    position_description,
    tech_stack,
    normalized_status,
    status_note,
    created_ts,
    updated_ts,
    import_note,
    import_note_created_at
)
with legacy_data (
    id,
    company,
    position_title,
    application_date,
    work_type,
    link,
    resolution,
    salary,
    position_description,
    tech_stack,
    comment,
    created_at,
    city,
    country
) as (
    values
        (4, 'Felmo', '', date '2026-02-07', 'On-Site', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.033908', 'Berlin', 'Germany'),
        (11, 'Meteor mobile', '', date '2026-02-26', 'Remote', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.043713', '', ''),
        (12, 'Company logo for, Jobgether.Jobgether', '', date '2026-02-26', 'Pending Info', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.045069', '', ''),
        (14, 'Academy Smart', '', date '2026-02-26', 'Remote', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.047579', '', ''),
        (17, 'DOMA Personal GmbH', '', date '2026-03-02', 'Remote', 'https://www.arbeitsagentur.de/jobsuche/jobdetail/10001-1002087732-S', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.051314', '', ''),
        (44, 'THRYVE', '', date '2026-03-12', 'Hybrid', '', 'Rejected', '80-100', '', 'go, golang', '', timestamp '2026-03-25 21:24:54.084309', '', ''),
        (21, 'Digisourced', '', date '2026-03-02', 'Pending Info', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.055938', '', ''),
        (24, 'Delivery Hero', '', date '2026-03-04', 'Remote', 'https://jobs.smartrecruiters.com/my-applications/DeliveryHero/8a879749-729b-4543-b016-dfa50c063745?dcr_ci=DeliveryHero', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.059366', '', ''),
        (28, 'Propel', '', date '2026-03-04', 'Pending Info', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.063950', '', ''),
        (36, 'Attio', '', date '2026-03-09', 'Remote', 'https://jobs.ashbyhq.com/attio/12b5faed-50b9-4b7a-80e8-af17b621cbe2?utm_source=LinkedInPaidJobWrapping', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.073903', '', ''),
        (40, 'Blackwall', '', date '2026-03-12', 'Remote', 'https://careers.blackwall.com/jobs/7297000-senior-php-engineer?promotion=1862186-trackable-share-link-li-jobs', 'Pending', '75-80', '', 'php, go', 'php, go', timestamp '2026-03-25 21:24:54.078828', '', ''),
        (41, 'Spectrum Search', '', date '2026-03-12', 'On-Site', '', 'Pending', '', '', 'php', 'FinTech PHP', timestamp '2026-03-25 21:24:54.080055', '', ''),
        (46, 'DocuMe', '', date '2026-03-16', 'Remote', '`', 'Pending', '60k', '', '', '', timestamp '2026-03-25 21:24:54.086741', '', ''),
        (48, 'Webit German Speaking', '', date '2026-03-16', 'Remote', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.089155', '', ''),
        (49, 'Zeelo', '', date '2026-03-16', 'Remote', '', 'Pending', '', '', 'php', 'PHP, Spain', timestamp '2026-03-25 21:24:54.090313', '', 'Spain'),
        (50, 'Planner 5D', '', date '2026-03-16', 'Remote', '', 'Pending', '', '', '', 'Remote', timestamp '2026-03-25 21:24:54.091556', 'Tallinn', 'Estonia'),
        (51, 'Kaufland', '', date '2026-03-16', 'Remote', 'https://kaufland-ecommerce.com/job/php-engineer-all-genders/', 'Pending', '75k', '', 'php, go', 'Remote PHP/Go', timestamp '2026-03-25 21:24:54.096303', '', 'Germany'),
        (18, 'Senior Backend Engineer - Golang (in Dresden)', '', date '2026-03-02', 'On-Site', 'https://staffbase.com/jobs/senior-backend-engineer-mfx-R579-dresden-8424464002/apply/?gh_jid=8424464002&gh_src=915a20462us', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.052473', 'Dresden', 'Germany'),
        (54, 'Code COmpass', '', date '2026-03-18', 'Pending Info', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.101495', '', ''),
        (55, 'Aalro', 'Senior PHP Engineer', date '2026-03-18', 'Remote', '', 'Pending', '', '', '', 'Remote Romain', timestamp '2026-03-25 21:24:54.102860', '', 'Romania'),
        (56, 'Internations', '', date '2026-03-24', 'Remote', 'https://internations.jobs.personio.de/job/99085', 'Pending', '80k', '', '', 'Remote, Munchen', timestamp '2026-03-25 21:24:54.104128', 'Munich', 'Germany'),
        (57, 'Propel', '', date '2026-03-24', 'Remote', '', 'Pending', '', '', 'symfony', 'Remote, Symfony, Russian', timestamp '2026-03-25 21:24:54.105333', '', ''),
        (58, 'Delivery Hero', '', date '2026-03-24', 'Remote', '', 'Pending', '', '', 'go', 'Remote, Quick COmmerce, Go', timestamp '2026-03-25 21:24:54.106540', '', ''),
        (59, 'Turing 42', '', date '2026-03-24', 'Remote', '', 'Pending', '', '', 'go', 'Remote, Go, Ab Sofort, Munchen', timestamp '2026-03-25 21:24:54.107792', 'Munich', 'Germany'),
        (60, 'Via', '', date '2026-03-24', 'Remote', 'https://www.vio.com/careers?ashby_jid=f70bf925-dc81-486f-ad1d-5178285cbb20', 'Pending', '', '', '', 'Remote, Amsterdam', timestamp '2026-03-25 21:24:54.109033', 'Amsterdam', 'Netherlands'),
        (3, 'DoctorLib', '', date '2026-02-07', 'On-Site', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.032173', 'Berlin', 'Germany'),
        (5, 'ottonova', '', date '2026-02-07', 'On-Site', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.035256', 'Munich', 'Germany'),
        (6, 'lumaserv', '', date '2026-02-07', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.036636', '', ''),
        (7, 'Kaufland', '', date '2026-02-07', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.037975', '', ''),
        (8, 'Kaufland', '', date '2026-02-26', 'Remote', 'attempt 2', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.039411', '', ''),
        (9, 'Flink', '', date '2026-02-26', 'Remote', 'https://jobs.smartrecruiters.com/my-applications/Flink3/20c7aeb8-cd35-4407-95bd-acedf43d9e1a?dcr_ci=Flink3', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.040731', '', ''),
        (13, 'Propel (revizto)', '', date '2026-02-26', 'Pending Info', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.046397', '', ''),
        (16, 'Revizto', '', date '2026-02-27', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.050093', '', ''),
        (19, 'Deel', '', date '2026-03-02', 'Remote', 'https://www.deel.com/careers/position/?ashby_jid=51d0d9a4-1f45-4225-ad1e-36b8beafb079&utm_source=Ob3JNy9ZxM', 'Rejected', '', '', 'node.js, kubernetes', 'Back-End/Infra Engineer (Kubernetes / Node.js)', timestamp '2026-03-25 21:24:54.053622', '', ''),
        (20, 'metiundo', '', date '2026-03-02', 'Pending Info', 'https://jobs.metiundo.io/Senior-Golang-Developer-mwx-de-f33.html', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.054793', '', ''),
        (22, 'Verisk', '', date '2026-03-04', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.057087', '', ''),
        (23, 'Jumingo', '', date '2026-03-04', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.058207', '', ''),
        (25, 'Sixt', '', date '2026-03-04', 'Pending Info', 'https://jobs.smartrecruiters.com/my-applications/SIXT/0137b846-4bfb-4e5e-8d6e-91945831e963?dcr_ci=SIXT', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.060503', '', ''),
        (26, 'FlixTrain', '', date '2026-03-04', 'Pending Info', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.061652', '', ''),
        (27, 'Qonto', '', date '2026-03-04', 'Pending Info', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.062825', '', ''),
        (29, 'Coins Pad', '', date '2026-03-04', 'Pending Info', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.065099', '', ''),
        (45, 'Local Brand X', '', date '2026-03-12', 'Remote', '', 'Pending', '', '', 'php', '', timestamp '2026-03-25 21:24:54.085535', '', ''),
        (31, 'Emma', '', date '2026-03-08', 'Pending Info', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.067589', '', ''),
        (32, 'Foodji', '', date '2026-03-08', 'Remote', 'https://foodji.recruitee.com/o/golang-developer-all-genders', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.068967', 'Munich', 'Germany'),
        (33, 'Wolt', '', date '2026-03-08', 'On-Site', 'https://job-boards.greenhouse.io/wolt/jobs/7682468', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.070152', '', ''),
        (34, 'EMEA', '', date '2026-03-09', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.071337', '', ''),
        (35, 'TYK', '', date '2026-03-09', 'Pending Info', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.072625', '', ''),
        (37, 'Zero 2 One Search', '', date '2026-03-09', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.075153', '', ''),
        (39, 'GLS Next', '', date '2026-03-12', 'Hybrid', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.077488', 'Berlin', 'Germany'),
        (42, 'Coins pad', '', date '2026-03-12', 'On-Site', 'https://jobs.eu.lever.co/coinspaid/e888f3b2-9c0c-45af-85a1-45d02a98344f/apply#', 'Rejected', '', '', '', 'russian + eng', timestamp '2026-03-25 21:24:54.081646', '', ''),
        (47, 'Pigment', '', date '2026-03-16', 'Remote', '', 'Rejected', '', '', '', '', timestamp '2026-03-25 21:24:54.087974', '', ''),
        (10, 'Matomo', '', date '2026-02-26', 'Remote', '', 'In Progress', '', '', '', 'Не сделал вторую задачу', timestamp '2026-03-25 21:24:54.042178', '', ''),
        (30, 'Hunty', '', date '2026-03-16', 'On-Site', '', 'Rejected', '', '', '', 'Poland', timestamp '2026-03-25 21:24:54.066303', '', 'Poland'),
        (53, 'SumUp', '', date '2026-03-24', 'On-Site', '', 'Pending', '', '', '', 'screening', timestamp '2026-03-25 21:24:54.100097', '', ''),
        (1, 'Monta', '', date '2026-02-07', 'On-Site', '', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.024263', 'Berlin', 'Germany'),
        (15, 'WebFleet', '', date '2026-02-24', 'Pending Info', 'https://bridgestone.wd5.myworkdayjobs.com/en-US/WF_External_Careers/userHome?shared_id=MGY0ZjM0MjEtOWNhZS00YmM2LWIyNzMtMDIzNGNiYTI1OTQ5', 'Pending', '', '', '', '', timestamp '2026-03-25 21:24:54.048809', '', ''),
        (43, 'RKD', '', date '2026-03-12', 'Remote', 'https://www.linkedin.com/jobs/view/4383782651/?trk=eml-email_application_confirmation_with_nba_01-applied_jobs-0-jobcard_body_4383782651&refId=fuwP1cj8TrOXiqsJ601OPg%3D%3D&trackingId=j9CxLqgnRmaKdO%2BIhUMePg%3D%3D', 'Pending', '', '', 'go, golang', 'GoLang', timestamp '2026-03-25 21:24:54.082980', '', ''),
        (38, 'Staffbase', '', date '2026-03-11', 'Remote', '', 'In Progress', '85k-95k', '', '', '', timestamp '2026-03-25 21:24:54.076302', 'Dresden', 'Germany'),
        (2, 'Amboss', '', date '2026-02-07', 'On-Site', '', 'Rejected after coding session', '', '', '', '', timestamp '2026-03-25 21:24:54.030361', 'Berlin', 'Germany'),
        (52, 'vialytics', '', date '2026-03-17', 'Remote', 'https://career.vialytics.com/en-GB/jobs/7392770-senior-backend-engineer-go-golang-m-f-d', 'Pending', '85k', '', 'go', 'Remote Go', timestamp '2026-03-25 21:24:54.098238', '', ''),
        (61, 'Pliant', 'Software Engineer - Backend', date '2026-03-26', 'Remote', '', 'Pending', '75000', 'Fintech B2B payment solutions. Modular, API-first platform.', 'Java, Spring Boot, REST APIs, PostgreSQL, AWS, Docker', 'Added via chat. Includes EU/UK remote.', timestamp '2026-03-26 14:32:06.586708', 'Berlin', 'Germany'),
        (62, 'refurbed', 'Senior Full-Stack Engineer - Catalyst Squad', date '2026-03-26', 'Remote', 'https://careers.refurbed.com/', 'Pending', '75000', 'Marketplace for refurbished products. Catalyst Squad.', 'Go, JavaScript, modern HTML/CSS, Tailwind, PostgreSQL, Redis, gRPC', 'Pasted along with the Pliant job in chat. Added just in case.', timestamp '2026-03-26 14:32:06.589470', 'Vienna', 'Austria')
)
select
    id,
    case
        when nullif(trim(position_title), '') is not null then nullif(trim(company), '')
        when company ~* '(engineer|developer|backend|full-stack|golang|software)' then null
        else nullif(trim(company), '')
    end as company_name,
    coalesce(
        nullif(trim(position_title), ''),
        case
            when company ~* '(engineer|developer|backend|full-stack|golang|software)' then nullif(trim(company), '')
        end,
        'Legacy application #' || id::text
    ) as normalized_position_title,
    case
        when nullif(trim(link), '') ~* '^https?://' then nullif(trim(link), '')
        else null
    end as source_url,
    nullif(trim(work_type), '') as normalized_work_type,
    nullif(trim(salary), '') as normalized_salary,
    nullif(trim(position_description), '') as normalized_position_description,
    nullif(trim(tech_stack), '') as normalized_tech_stack,
    case
        when lower(trim(resolution)) in ('rejected', 'rejected after coding session') then 'rejected'
        when lower(trim(resolution)) in ('in progress', 'pending info') then 'screening'
        else 'applied'
    end as normalized_status,
    case
        when lower(trim(resolution)) in ('pending', 'rejected', '') then null
        else 'Imported legacy resolution: ' || trim(resolution)
    end as status_note,
    coalesce(application_date::timestamp, created_at)::timestamptz as created_ts,
    greatest(coalesce(application_date::timestamp, created_at), created_at)::timestamptz as updated_ts,
    trim(both E'\n' from concat_ws(
        E'\n',
        'Imported from legacy dataset.',
        case
            when application_date is not null then 'Legacy application date: ' || application_date::text
        end,
        case
            when nullif(trim(city), '') is not null or nullif(trim(country), '') is not null then
                'Legacy location: ' || concat_ws(', ', nullif(trim(city), ''), nullif(trim(country), ''))
        end,
        case
            when nullif(trim(comment), '') is not null then 'Legacy note: ' || trim(comment)
        end,
        case
            when nullif(trim(link), '') is not null and nullif(trim(link), '') !~* '^https?://' then
                'Legacy source value: ' || trim(link)
        end
    )) as import_note,
    created_at::timestamptz as import_note_created_at
from legacy_data;

insert into public.job_applications (
    id,
    company_name,
    position_title,
    source_url,
    work_type,
    salary,
    position_description,
    tech_stack,
    created_at,
    updated_at
)
select
    id,
    company_name,
    position_title,
    source_url,
    work_type,
    salary,
    position_description,
    tech_stack,
    created_ts,
    updated_ts
from tmp_legacy_job_applications_import
on conflict (id) do nothing;

insert into public.job_application_status_history (
    application_id,
    status,
    note,
    changed_at
)
select
    i.id,
    i.normalized_status,
    i.status_note,
    i.created_ts
from tmp_legacy_job_applications_import i
where not exists (
    select 1
    from public.job_application_status_history sh
    where sh.application_id = i.id
);

insert into public.job_application_comments (
    application_id,
    body,
    created_at
)
select
    i.id,
    i.import_note,
    i.import_note_created_at
from tmp_legacy_job_applications_import i
where i.import_note <> ''
  and not exists (
      select 1
      from public.job_application_comments c
      where c.application_id = i.id
        and c.body = i.import_note
  );

select setval(
    pg_get_serial_sequence('public.job_applications', 'id'),
    coalesce((select max(id) from public.job_applications), 1),
    true
);
