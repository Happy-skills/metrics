CREATE TABLE metric_values (
    id integer generated always as identity primary key,
    name text not null,
    type text not null,
    value float8,
    delta int8
);

CREATE INDEX idx_metric_name ON metric_values(name);

ALTER TABLE metric_values ADD CONSTRAINT metric_values_unique_name UNIQUE (name, type);