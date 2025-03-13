CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE check_definitions
(
    filename                               text    not null PRIMARY KEY,
    check_command                          text    not null,
    execute_on_failure                     text             default null,
    execute_on_timeout                     text             default null,
    interval_seconds_between_checks        integer not null CHECK (interval_seconds_between_checks BETWEEN 5 AND 3600),
    delay_seconds_before_first_check       integer not null CHECK (delay_seconds_before_first_check BETWEEN 0 AND 600),
    timeout_seconds                        integer not null CHECK (timeout_seconds BETWEEN 1 AND 120),
    stop_checking_after_number_of_timeouts integer not null CHECK (stop_checking_after_number_of_timeouts BETWEEN 1 AND 10),
    last_run_timestamp                     integer not null default 0
) strict;
CREATE INDEX idx_check_definitions_filename ON check_definitions (filename);
CREATE TABLE results
(
    filename text,
    rc       integer not null check (rc BETWEEN 0 and 255),
    name     text not null,
    text     text default null,
    perfdata text default null,
    host     text default null,
    tags     text default null,
    foreign key (filename) references check_definitions(filename) on delete cascade on update cascade
) strict;
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20250106102647');
