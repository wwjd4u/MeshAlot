BEGIN;

CREATE TABLE node_ratings(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    node_id uuid NOT NULL
        REFERENCES nodes(id),

    compute_score integer,
    network_score integer,
    reliability_score integer,
    availability_score integer,
    trust_score integer,

    compute_known boolean NOT NULL,
    network_known boolean NOT NULL,
    reliability_known boolean NOT NULL,
    availability_known boolean NOT NULL,
    trust_known boolean NOT NULL,

    tier text NOT NULL,

    observed_at timestamptz NOT NULL,

    CHECK (
        compute_score IS NULL OR
        compute_score BETWEEN 0 AND 100
    ),

    CHECK (
        network_score IS NULL OR
        network_score BETWEEN 0 AND 100
    ),

    CHECK (
        reliability_score IS NULL OR
        reliability_score BETWEEN 0 AND 100
    ),

    CHECK (
        availability_score IS NULL OR
        availability_score BETWEEN 0 AND 100
    ),

    CHECK (
        trust_score IS NULL OR
        trust_score BETWEEN 0 AND 100
    ),

    CHECK (
        tier IN (
            'unrated',
            'basic',
            'standard',
            'gold',
            'platinum',
            'datacenter'
        )
    ),

    CHECK (
        (compute_known AND compute_score IS NOT NULL)
        OR
        (NOT compute_known AND compute_score IS NULL)
    ),

    CHECK (
        (network_known AND network_score IS NOT NULL)
        OR
        (NOT network_known AND network_score IS NULL)
    ),

    CHECK (
        (reliability_known AND reliability_score IS NOT NULL)
        OR
        (NOT reliability_known AND reliability_score IS NULL)
    ),

    CHECK (
        (availability_known AND availability_score IS NOT NULL)
        OR
        (NOT availability_known AND availability_score IS NULL)
    ),

    CHECK (
        (trust_known AND trust_score IS NOT NULL)
        OR
        (NOT trust_known AND trust_score IS NULL)
    ),

    CHECK (
        (
            compute_known
            AND network_known
            AND reliability_known
            AND availability_known
            AND trust_known
            AND tier <> 'unrated'
        )
        OR
        (
            NOT (
                compute_known
                AND network_known
                AND reliability_known
                AND availability_known
                AND trust_known
            )
            AND tier = 'unrated'
        )
    )
);

CREATE INDEX node_ratings_node_observed_idx
    ON node_ratings(node_id, observed_at DESC);

COMMIT;
