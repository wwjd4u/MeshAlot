\set ON_ERROR_STOP on

BEGIN;

DO $$
DECLARE
    function_oid oid;
    function_is_security_definer boolean;
BEGIN
    IF current_database() !~ '^meshalot(_[a-z0-9]+)*$' THEN
        RAISE EXCEPTION
            'Refusing M8 runtime privilege test outside a MeshAlot database';
    END IF;

    SELECT
        p.oid,
        p.prosecdef
    INTO
        function_oid,
        function_is_security_definer
    FROM pg_proc p
    WHERE p.oid =
        'public.insert_network_benchmark(uuid,text,jsonb,numeric,timestamptz)'
        ::regprocedure;

    IF function_oid IS NULL THEN
        RAISE EXCEPTION
            'M8 network benchmark insert function is missing';
    END IF;

    IF NOT function_is_security_definer THEN
        RAISE EXCEPTION
            'M8 network benchmark insert function is not SECURITY DEFINER';
    END IF;

    IF NOT has_function_privilege(
        'meshalot',
        'public.insert_network_benchmark(uuid,text,jsonb,numeric,timestamptz)',
        'EXECUTE'
    ) THEN
        RAISE EXCEPTION
            'meshalot runtime role cannot execute M8 benchmark insert function';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_proc p
        CROSS JOIN LATERAL
            aclexplode(
                COALESCE(
                    p.proacl,
                    acldefault('f', p.proowner)
                )
            ) acl
        WHERE p.oid = function_oid
          AND acl.grantee = 0
          AND acl.privilege_type = 'EXECUTE'
    ) THEN
        RAISE EXCEPTION
            'PUBLIC can execute M8 benchmark insert function';
    END IF;

    IF has_table_privilege(
        'meshalot',
        'network_benchmarks',
        'INSERT'
    )
       OR has_table_privilege(
           'meshalot',
           'network_benchmarks',
           'UPDATE'
       )
       OR has_table_privilege(
           'meshalot',
           'network_benchmarks',
           'DELETE'
       ) THEN
        RAISE EXCEPTION
            'meshalot runtime role has direct network benchmark write rights';
    END IF;

    IF has_table_privilege(
        'meshalot',
        'compute_benchmarks',
        'INSERT'
    )
       OR has_table_privilege(
           'meshalot',
           'compute_benchmarks',
           'UPDATE'
       )
       OR has_table_privilege(
           'meshalot',
           'compute_benchmarks',
           'DELETE'
       ) THEN
        RAISE EXCEPTION
            'meshalot runtime role has direct compute benchmark write rights';
    END IF;
END
$$;

SELECT
    'MILESTONE 8 RUNTIME PRIVILEGE ASSERTIONS PASSED'
    AS result;

ROLLBACK;
