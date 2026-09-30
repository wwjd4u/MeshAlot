\set ON_ERROR_STOP on

BEGIN;

DO $$
DECLARE
    function_oid oid;
    function_is_security_definer boolean;
BEGIN
    IF current_database() !~ '^meshalot(_[a-z0-9]+)*$' THEN
        RAISE EXCEPTION
            'Refusing M10 runtime privilege test outside a MeshAlot database';
    END IF;

    SELECT
        p.oid,
        p.prosecdef
    INTO
        function_oid,
        function_is_security_definer
    FROM pg_proc p
    WHERE p.oid = to_regprocedure(
        'public.insert_node_rating(text,integer,integer,integer,integer,integer,boolean,boolean,boolean,boolean,boolean,text,timestamptz)'
    );

    IF function_oid IS NULL THEN
        RAISE EXCEPTION
            'M10 node rating insert function is missing';
    END IF;

    IF NOT function_is_security_definer THEN
        RAISE EXCEPTION
            'M10 node rating insert function is not SECURITY DEFINER';
    END IF;

    IF NOT has_function_privilege(
        'meshalot',
        'public.insert_node_rating(text,integer,integer,integer,integer,integer,boolean,boolean,boolean,boolean,boolean,text,timestamptz)',
        'EXECUTE'
    ) THEN
        RAISE EXCEPTION
            'meshalot runtime role cannot execute M10 node rating insert function';
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
            'PUBLIC can execute M10 node rating insert function';
    END IF;

    IF has_table_privilege(
        'meshalot',
        'node_ratings',
        'INSERT'
    )
       OR has_table_privilege(
           'meshalot',
           'node_ratings',
           'UPDATE'
       )
       OR has_table_privilege(
           'meshalot',
           'node_ratings',
           'DELETE'
       ) THEN
        RAISE EXCEPTION
            'meshalot runtime role has direct node rating write rights';
    END IF;

    IF NOT has_table_privilege(
        'meshalot',
        'node_ratings',
        'SELECT'
    ) THEN
        RAISE EXCEPTION
            'meshalot runtime role cannot read node ratings';
    END IF;
END
$$;

SELECT
    'MILESTONE 10 RUNTIME PRIVILEGE ASSERTIONS PASSED'
    AS result;

ROLLBACK;
