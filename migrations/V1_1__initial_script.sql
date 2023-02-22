create table if not exists position
(
    id          varchar(200)
        constraint Position_pk
            primary key,
    user        varchar(200)                              not null,
    name        varchar(50)                               not null,
    amount      double precision                          not null,
    stop_loss   double precision,
    take_profit double precision,
    closed      timestamp(6),
    created     timestamp(6) default CURRENT_TIMESTAMP(6) not null,
    updated     timestamp(6) default CURRENT_TIMESTAMP(6) not null
);

alter table position
    owner to postgres;

create unique index if not exists position_user_name_closed_uindex
    on position (user, name, closed);

CREATE TRIGGER stop_loss
    AFTER UPDATE OF stop_loss
    ON position
EXECUTE FUNCTION notify();

CREATE TRIGGER take_profit
    AFTER UPDATE OF take_profit
    ON position
EXECUTE FUNCTION notify();

CREATE OR REPLACE FUNCTION notify() RETURNS TRIGGER AS
$FN$
DECLARE
    payload jsonb;
BEGIN
    payload = jsonb_build_object(
            'id', to_jsonb(OLD.id),
            'name', to_jsonb(OLD.name),
            'type', to_jsonb(TG_NAME),
        );
    IF TG_NAME = 'stop_loss' THEN
        payload = jsonb_set(payload, 'amount', OLD.stop_loss, true);
    END IF;
    IF TG_NAME = 'take_profit' THEN
        payload = jsonb_set(payload, 'amount', OLD.take_profit, true);
    END IF;

    PERFORM pg_notify('thresholds', payload::TEXT);

    RETURN NEW;
END;
$FN$ LANGUAGE plpgsql;