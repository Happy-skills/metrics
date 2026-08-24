DROP INDEX IF EXISTS idx_metric_name;

ALTER TABLE metric_values DROP CONSTRAINT metric_values_unique_name;

DROP TABLE IF EXISTS metric_values;
