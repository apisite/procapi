/*
  Functions for testing args

*/

-- -----------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION test_args (
  a_name TEXT DEFAULT 'def'
) RETURNS TEXT STABLE LANGUAGE 'sql' AS
$_$
  SELECT a_name
$_$;

-- -----------------------------------------------------------------------------
