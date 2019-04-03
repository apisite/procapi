/*
  Mock functions for pgfc testing.
  Real example: see https://github.com/pomasql/rpc

*/

\set WD `printenv WORKDIR`

\set ID rpc
-- Create schema
CREATE SCHEMA IF NOT EXISTS :ID;
\set QUIET on
SET SEARCH_PATH = :ID, public;
\set QUIET off


CREATE TYPE func_def AS (
  code      text
, nspname   name
, proname   name
, is_set    boolean
, is_ro     boolean
, is_struct boolean
, result    text
, anno      text
, sample    text
);

CREATE OR REPLACE FUNCTION create_index(a_json TEXT) RETURNS VOID LANGUAGE plpgsql AS $_$
BEGIN
  execute format(
    $__$CREATE OR REPLACE FUNCTION index(a_nsp TEXT DEFAULT NULL) RETURNS SETOF func_def STABLE LANGUAGE sql AS
     $$ select * from json_populate_recordset(null::func_def, %L)
        WHERE a_nsp IS NULL OR a_nsp = current_schema();$$;
    $__$, a_json
  );
END;
$_$;

CREATE OR REPLACE FUNCTION create_func(a_name TEXT, a_type TEXT, a_json TEXT) RETURNS VOID LANGUAGE plpgsql AS $_$
BEGIN
  execute format(
    $__$CREATE OR REPLACE FUNCTION %I(a_code TEXT) RETURNS SETOF %I STABLE LANGUAGE sql AS
      $$ select * from jsonb_populate_recordset(null::%I, %L::JSONB -> a_code)$$;
    $__$, a_name, a_type, a_type, a_json
  );
END;
$_$;

-- \set CMD 'jq -c \'.\' < ':WD /index.json
\set CMD 'cat ':WD /index.json
\set JSON `:CMD`
SELECT create_index(:'JSON');

SELECT * FROM index();
SELECT * FROM index(:'ID');
SELECT * FROM index('ddd');

CREATE TYPE arg_def AS (
  arg     TEXT
, type     TEXT
, required BOOL
, def_val  TEXT
, anno     TEXT
); -- ? INHERITS(result_def)


--\set JSON `jq -c '.' < args.json`
\set CMD 'cat ':WD /args.json
\set JSON `:CMD`
SELECT create_func('func_args','arg_def',:'JSON');

select * from func_args('index');
select * from func_args('func_args');
select * from func_args('func_result');
select * from func_args('unknown');


CREATE TYPE result_def AS (
  arg TEXT
, type TEXT
, anno TEXT
);

--\set JSON `jq -c '.' < result.json`
\set CMD 'cat ':WD /result.json
\set JSON `:CMD`

SELECT create_func('func_result','result_def',:'JSON');


select * from func_result('index');
select * from func_result('func_args');
select * from func_result('func_result');
select * from func_result('unknown');

/*
  Testing funcs
*/

\set PKG rpc_testing
CREATE SCHEMA :PKG;
SET SEARCH_PATH = :PKG, 'public';

\set SQL :WD /50_args.sql
\i :SQL
\set SQL :WD /50_types.sql
\i :SQL
