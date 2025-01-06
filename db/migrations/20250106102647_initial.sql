-- migrate:up
create table check_definitions
(
    filename                               text    not null,
    check_command                          text    not null,
    interval_seconds_between_checks        integer not null CHECK (interval_seconds_between_checks BETWEEN 5 AND 3600),
    delay_seconds_before_first_check       integer not null CHECK (delay_seconds_before_first_check BETWEEN 0 AND 600),
    timeout_seconds                        integer not null CHECK (timeout_seconds BETWEEN 1 AND 120),
    stop_checking_after_number_of_timeouts integer not null CHECK (stop_checking_after_number_of_timeouts BETWEEN 1 AND 10)
) strict;

CREATE INDEX idx_check_definitions_filename ON check_definitions (filename);

-- migrate:down

drop table check_definitions;