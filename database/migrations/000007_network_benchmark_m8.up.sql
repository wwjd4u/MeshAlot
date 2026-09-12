BEGIN;

CREATE INDEX network_benchmarks_node_observed_idx
    ON network_benchmarks(node_id, observed_at DESC);

CREATE FUNCTION public.insert_network_benchmark(
    p_report_id uuid,
    p_node_key text,
    p_payload jsonb,
    p_score numeric,
    p_observed_at timestamptz
)
RETURNS integer
LANGUAGE sql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    WITH inserted AS (
        INSERT INTO public.network_benchmarks(
            id,
            node_id,
            payload,
            score,
            observed_at
        )
        SELECT
            p_report_id,
            n.id,
            p_payload,
            p_score,
            p_observed_at
        FROM public.nodes n
        WHERE n.node_key = p_node_key
        RETURNING 1
    )
    SELECT count(*)::integer
    FROM inserted;
$$;

REVOKE ALL
ON FUNCTION public.insert_network_benchmark(
    uuid,
    text,
    jsonb,
    numeric,
    timestamptz
)
FROM PUBLIC;

COMMIT;
