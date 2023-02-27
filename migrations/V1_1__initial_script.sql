create table if not exists positions
(
    id             varchar(200)
        constraint Positions_pk
            primary key,
    "user"         varchar(200)                                  not null,
    "name"         varchar(50)                                   not null,
    amount         double precision                              not null,
    stop_loss      double precision default 0                    not null,
    take_profit    double precision default 0                    not null,
    short_position double precision default 0                    not null,
    closed         bigint           default 0                    not null,
    created        timestamp(6)     default CURRENT_TIMESTAMP(6) not null,
    updated        timestamp(6)     default CURRENT_TIMESTAMP(6) not null
);

alter table positions
    owner to postgres;

create UNIQUE index if not exists positions_user_name_closed_uindex
    on positions ("user", "name", closed);

CREATE OR REPLACE FUNCTION notify() RETURNS TRIGGER AS
$BODY$
DECLARE
    payload jsonb;
BEGIN
    payload = jsonb_build_object(
            'id', to_jsonb(OLD.id),
            'name', to_jsonb(OLD.name),
            'user', to_jsonb(OLD.user),
            'short_position', to_jsonb(OLD.short_position),
            'type', to_jsonb(TG_NAME)
        );
    IF TG_NAME = 'stop_loss' THEN
        payload = jsonb_set(payload, '{stop_loss}', to_jsonb(OLD.stop_loss), true);
    END IF;
    IF TG_NAME = 'take_profit' THEN
        payload = jsonb_set(payload, '{take_profit}', to_jsonb(OLD.take_profit), true);
    END IF;
    IF TG_NAME = 'closed' THEN
        payload = jsonb_set(payload, '{closed}', to_jsonb(NEW.closed), true);
        payload = jsonb_set(payload, '{stop_loss}', to_jsonb(OLD.stop_loss), true);
        payload = jsonb_set(payload, '{take_profit}', to_jsonb(OLD.take_profit), true);
    END IF;
    IF TG_NAME = 'created' THEN
        payload = jsonb_set(payload, '{amount}', to_jsonb(NEW.amount), true);
    END IF;

    PERFORM pg_notify('thresholds', payload::TEXT);

    RETURN NEW;
END ;
$BODY$ LANGUAGE plpgsql;

CREATE TRIGGER stop_loss
    AFTER UPDATE OF stop_loss
    ON positions
    FOR EACH ROW
EXECUTE FUNCTION notify();

CREATE TRIGGER take_profit
    AFTER UPDATE OF take_profit
    ON positions
    FOR EACH ROW
EXECUTE FUNCTION notify();

CREATE TRIGGER closed
    AFTER UPDATE OF closed
    ON positions
    FOR EACH ROW
EXECUTE FUNCTION notify();

CREATE TRIGGER created
    AFTER INSERT
    ON positions
    FOR EACH ROW
EXECUTE FUNCTION notify();