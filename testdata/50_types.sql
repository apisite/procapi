/*
  Functions for testing types

TODO: move to table

Tested types:

bool
char
date
float4
float8
inet
int2
int4
int8
interval
json
jsonb
money
numeric
text 
time
timestamp
timestamptz
int4[]
text[]

*/

-- -----------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION test_types (
  a_tbool bool              
, a_tchar char
, a_tdate date
, a_tfloat4 float4
, a_tfloat8 float8
, a_tinet inet
, a_tint2 int2
, a_tint4 int4
, a_tint8 int8
, a_tinterval interval
, a_tjson json
, a_tjsonb jsonb
, a_tmoney money
, a_tnumeric numeric
, a_ttext  text 
, a_ttime time
, a_ttimestamp timestamp
, a_ttimestamptz timestamptz
, a_aint4 int4[]
, a_atext text[]
) RETURNS TABLE(
  id int
, tbool bool              
, tchar char
, tdate date
, tfloat4 float4
, tfloat8 float8
, tinet inet
, tint2 int2
, tint4 int4
, tint8 int8
, tinterval interval
, tjson json
, tjsonb jsonb
, tmoney money
, tnumeric numeric
, ttext  text 
, ttime time
, ttimestamp timestamp
, ttimestamptz timestamptz
, aint4 int4[]
, atext text[]
) STABLE LANGUAGE 'sql' AS
$_$
  SELECT
  1::int 
, a_tbool
, a_tchar
, a_tdate
, a_tfloat4
, a_tfloat8
, a_tinet
, a_tint2
, a_tint4
, a_tint8
, a_tinterval
, a_tjson
, a_tjsonb
, a_tmoney
, a_tnumeric
, a_ttext 
, a_ttime
, a_ttimestamp
, a_ttimestamptz
, a_aint4
, a_atext
UNION ALL SELECT
  2::int
, NOT a_tbool
, 'x'::char
, '2019-03-31'::date
, (a_tfloat4 / 3)::float4
, (a_tfloat8 / 3)::float8
, a_tinet
, (a_tint2 / 2)::int2
, (a_tint4 / 2)::int4
, (a_tint8 / 2)::int8
, a_tinterval + '1 month'::interval
, a_tjson
, a_tjsonb
, a_tmoney
, a_tnumeric
, a_ttext  || a_ttext
, '23:55:10.50'::time
, '12/17/1997 15:37:16.00'::timestamp
, '1997-12-17 12:00 EDT'::timestamptz
, array[9,8,7]::int4[]
, array['zyx1','zyx2']::text[]
UNION ALL SELECT
  3::int
, NULL::bool
, NULL::char
, NULL::date
, NULL::float4
, NULL::float8
, NULL::inet
, NULL::int2
, NULL::int4
, NULL::int8
, NULL::interval
, NULL::json
, NULL::jsonb
, NULL::money
, NULL::numeric
, NULL::text 
, NULL::time
, NULL::timestamp
, NULL::timestamptz
, NULL::int4[]
, NULL::text[]
$_$;
