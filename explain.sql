INSERT INTO orders (id, client_id, pickup_point_id, status, status_updated, weight, cost)
SELECT
    gs AS id,
    (random() * 10000)::bigint AS client_id,
        (random() * 1000)::bigint AS pickup_point_id,
        (ARRAY['stored', 'reached-client', 'returned', 'canceled'])[floor(random() * 4 + 1)]::varchar AS status,
    NOW() - interval '1 day' * (random() * 365)::int AS status_updated,
    (random() * 100)::bigint AS weight,
    (random() * 1000)::bigint AS cost
FROM generate_series(1, 1000000) AS gs;
-- сгенерировал миллион строк в бд

-- по очереди добавлял индексы, смотрел результаты
CREATE INDEX CONCURRENTLY idx_orders_cid_hash
    ON orders USING hash (client_id);
CREATE INDEX CONCURRENTLY idx_orders_ppid_hash
    ON orders USING hash (pickup_point_id);
CREATE INDEX CONCURRENTLY idx_orders_status_hash
    ON orders USING hash (status);
--
-- drop'нул все hash индексы
CREATE INDEX CONCURRENTLY idx_orders_status_btree ON orders (status);
drop index idx_orders_status_btree;
--
CREATE INDEX CONCURRENTLY idx_orders_id_cid_ppid_status_btree ON orders (id, client_id, pickup_point_id, status);
drop index idx_orders_id_cid_ppid_status_btree;
--

EXPLAIN (ANALYSE, verbose, BUFFERS)
SELECT * FROM orders WHERE id = 5208;

-- результаты нескольких запусков

-- NO INDEX
-- Index Scan using orders_pkey on public.orders  (cost=0.42..8.44 rows=1 width=58) (actual time=0.059..0.062 rows=1 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.id = 256)
--   Buffers: shared hit=4
-- Planning Time: 0.132 ms
-- Execution Time: 0.205 ms
--
-- Index Scan using orders_pkey on public.orders  (cost=0.42..8.44 rows=1 width=58) (actual time=0.079..0.083 rows=1 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.id = 256)
--   Buffers: shared hit=4
-- Planning Time: 0.678 ms
-- Execution Time: 0.262 ms
--
-- Index Scan using orders_pkey on public.orders  (cost=0.42..8.44 rows=1 width=58) (actual time=0.356..0.360 rows=1 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.id = 108)
--   Buffers: shared hit=4
-- Planning Time: 0.105 ms
-- Execution Time: 0.426 ms

-- HASH id
-- Index Scan using idx_orders_id_hash on public.orders  (cost=0.00..8.02 rows=1 width=58) (actual time=0.565..0.911 rows=1 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.id = 179)
--   Buffers: shared hit=3
-- Planning Time: 0.197 ms
-- Execution Time: 0.958 ms

-- BTREE (id, cid, ppid, status)
-- Index Scan using idx_orders_id_cid_ppid_status_btree on public.orders  (cost=0.42..8.44 rows=1 width=58) (actual time=0.063..0.064 rows=1 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.id = 5508)
--   Buffers: shared hit=4
-- Planning Time: 0.079 ms
-- Execution Time: 0.102 ms
--
-- Index Scan using idx_orders_id_cid_ppid_status_btree on public.orders  (cost=0.42..8.44 rows=1 width=58) (actual time=0.093..0.095 rows=1 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.id = 5208)
--   Buffers: shared hit=4
-- Planning Time: 0.085 ms
-- Execution Time: 0.135 ms


EXPLAIN (ANALYSE, verbose, BUFFERS)
SELECT * FROM orders WHERE client_id = 727 AND pickup_point_id = 597;

-- NO INDEX
-- Gather  (cost=1000.00..18333.10 rows=1 width=58) (actual time=1031.459..1044.600 rows=0 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Workers Planned: 2
--   Workers Launched: 2
--   Buffers: shared hit=11083
--   ->  Parallel Seq Scan on public.orders  (cost=0.00..17333.00 rows=1 width=58) (actual time=974.623..974.624 rows=0 loops=3)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.client_id = 3) AND (orders.pickup_point_id = 5))
--         Rows Removed by Filter: 333333
--         Buffers: shared hit=11083
--         Worker 0:  actual time=954.456..954.458 rows=0 loops=1
--           Buffers: shared hit=3669
--         Worker 1:  actual time=954.570..954.570 rows=0 loops=1
--           Buffers: shared hit=3633
-- Planning Time: 0.519 ms
-- Execution Time: 1044.918 ms
--
-- Gather  (cost=1000.00..18333.10 rows=1 width=58) (actual time=138.214..145.291 rows=0 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Workers Planned: 2
--   Workers Launched: 2
--   Buffers: shared hit=11083
--   ->  Parallel Seq Scan on public.orders  (cost=0.00..17333.00 rows=1 width=58) (actual time=123.716..123.717 rows=0 loops=3)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.client_id = 10) AND (orders.pickup_point_id = 16))
--         Rows Removed by Filter: 333333
--         Buffers: shared hit=11083
--         Worker 0:  actual time=117.210..117.212 rows=0 loops=1
--           Buffers: shared hit=3137
--         Worker 1:  actual time=116.711..116.712 rows=0 loops=1
--           Buffers: shared hit=3196
-- Planning Time: 0.168 ms
-- Execution Time: 145.349 ms
--
-- Gather  (cost=1000.00..18333.10 rows=1 width=58) (actual time=129.315..149.085 rows=0 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Workers Planned: 2
--   Workers Launched: 2
--   Buffers: shared hit=11083
--   ->  Parallel Seq Scan on public.orders  (cost=0.00..17333.00 rows=1 width=58) (actual time=117.097..117.098 rows=0 loops=3)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.client_id = 15) AND (orders.pickup_point_id = 10))
--         Rows Removed by Filter: 333333
--         Buffers: shared hit=11083
--         Worker 0:  actual time=110.246..110.247 rows=0 loops=1
--           Buffers: shared hit=3288
--         Worker 1:  actual time=112.765..112.767 rows=0 loops=1
--           Buffers: shared hit=3283
-- Planning Time: 0.241 ms
-- Execution Time: 149.165 ms

-- HASH (cid, ppid)
-- Index Scan using idx_orders_cid_hash on public.orders  (cost=0.00..12.04 rows=1 width=58) (actual time=4.771..4.772 rows=0 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.client_id = 344)
--   Filter: (orders.pickup_point_id = 277)
--   Rows Removed by Filter: 3
--   Buffers: shared hit=1 read=3
-- Planning Time: 0.262 ms
-- Execution Time: 4.829 ms
--
-- Index Scan using idx_orders_cid_hash on public.orders  (cost=0.00..12.04 rows=1 width=58) (actual time=1.009..1.010 rows=0 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Index Cond: (orders.client_id = 787)
--   Filter: (orders.pickup_point_id = 888)
--   Buffers: shared hit=1
-- Planning:
--   Buffers: shared hit=5
-- Planning Time: 0.918 ms
-- Execution Time: 1.063 ms

-- BTREE (id, cid, ppid, status)
-- Gather  (cost=1000.00..18333.10 rows=1 width=58) (actual time=53.176..55.769 rows=0 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Workers Planned: 2
--   Workers Launched: 2
--   Buffers: shared hit=2805 read=8278
--   ->  Parallel Seq Scan on public.orders  (cost=0.00..17333.00 rows=1 width=58) (actual time=47.232..47.233 rows=0 loops=3)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.client_id = 77) AND (orders.pickup_point_id = 797))
--         Rows Removed by Filter: 333333
--         Buffers: shared hit=2805 read=8278
--         Worker 0:  actual time=44.816..44.817 rows=0 loops=1
--           Buffers: shared hit=759 read=2389
--         Worker 1:  actual time=45.986..45.987 rows=0 loops=1
--           Buffers: shared hit=868 read=2629
-- Planning Time: 0.088 ms
-- Execution Time: 55.803 ms


EXPLAIN (ANALYSE, verbose, BUFFERS)
SELECT * FROM orders WHERE client_id = 257 AND status = 'returned';
-- HASH (cid)
-- Bitmap Heap Scan on public.orders  (cost=4.75..374.16 rows=25 width=58) (actual time=0.685..2.041 rows=27 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Recheck Cond: (orders.client_id = 3)
--   Filter: ((orders.status)::text = 'stored'::text)
--   Rows Removed by Filter: 74
--   Heap Blocks: exact=101
--   Buffers: shared hit=94 read=9
--   ->  Bitmap Index Scan on idx_orders_cid_hash  (cost=0.00..4.74 rows=99 width=0) (actual time=0.460..0.462 rows=101 loops=1)
--         Index Cond: (orders.client_id = 3)
--         Buffers: shared hit=2
-- Planning:
--   Buffers: shared hit=49
-- Planning Time: 5.549 ms
-- Execution Time: 2.273 ms

-- HASH (cid, ppid)
-- Bitmap Heap Scan on public.orders  (cost=4.75..374.16 rows=25 width=58) (actual time=0.860..7.232 rows=27 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Recheck Cond: (orders.client_id = 977)
--   Filter: ((orders.status)::text = 'returned'::text)
--   Rows Removed by Filter: 69
--   Heap Blocks: exact=96
--   Buffers: shared hit=49 read=49
--   ->  Bitmap Index Scan on idx_orders_cid_hash  (cost=0.00..4.74 rows=99 width=0) (actual time=0.557..0.558 rows=96 loops=1)
--         Index Cond: (orders.client_id = 977)
--         Buffers: shared hit=2
-- Planning Time: 0.169 ms
-- Execution Time: 7.517 ms

-- HASH (cid, ppid, status)
-- Bitmap Heap Scan on public.orders  (cost=4.75..374.16 rows=25 width=58) (actual time=0.782..1.175 rows=24 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Recheck Cond: (orders.client_id = 777)
--   Filter: ((orders.status)::text = 'stored'::text)
--   Rows Removed by Filter: 68
--   Heap Blocks: exact=91
--   Buffers: shared hit=92
--   ->  Bitmap Index Scan on idx_orders_cid_hash  (cost=0.00..4.74 rows=99 width=0) (actual time=0.734..0.734 rows=92 loops=1)
--         Index Cond: (orders.client_id = 777)
--         Buffers: shared hit=1
-- Planning Time: 0.154 ms
-- Execution Time: 1.227 ms

-- HASH (cid, ppid) + BTREE(status)
-- Bitmap Heap Scan on public.orders  (cost=4.75..374.16 rows=25 width=58) (actual time=0.240..4.480 rows=19 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Recheck Cond: (orders.client_id = 257)
--   Filter: ((orders.status)::text = 'returned'::text)
--   Rows Removed by Filter: 73
--   Heap Blocks: exact=92
--   Buffers: shared hit=26 read=67
--   ->  Bitmap Index Scan on idx_orders_cid_hash  (cost=0.00..4.74 rows=99 width=0) (actual time=0.150..0.151 rows=92 loops=1)
--         Index Cond: (orders.client_id = 257)
--         Buffers: shared hit=1
-- Planning Time: 0.072 ms
-- Execution Time: 4.540 ms


EXPLAIN (ANALYSE, verbose, BUFFERS)
SELECT * FROM orders WHERE status = 'returned' LIMIT 58 OFFSET 789;
-- HASH (cid, ppid)
-- Limit  (cost=0.48..1.43 rows=10 width=58) (actual time=0.464..0.491 rows=10 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Buffers: shared read=1
--   ->  Seq Scan on public.orders  (cost=0.00..23583.00 rows=248233 width=58) (actual time=0.451..0.479 rows=15 loops=1)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.status)::text = 'returned'::text)
--         Rows Removed by Filter: 32
--         Buffers: shared read=1
-- Planning Time: 0.612 ms
-- Execution Time: 0.742 ms
--
-- Buffers: shared hit=23
--   ->  Seq Scan on public.orders  (cost=0.00..23583.00 rows=251100 width=58) (actual time=0.093..4.241 rows=499 loops=1)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.status)::text = 'reached-client'::text)
--         Rows Removed by Filter: 1565
--         Buffers: shared hit=23
-- Planning Time: 0.201 ms
-- Execution Time: 4.428 ms
--
-- Buffers: shared hit=22
--   ->  Seq Scan on public.orders  (cost=0.00..23583.00 rows=249300 width=58) (actual time=0.020..3.007 rows=499 loops=1)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.status)::text = 'canceled'::text)
--         Rows Removed by Filter: 1395
--         Buffers: shared hit=22
-- Planning Time: 0.130 ms
-- Execution Time: 3.131 ms

-- HASH (cid, ppid, status)
-- Limit  (cost=45.51..47.41 rows=20 width=58) (actual time=0.378..0.389 rows=20 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Buffers: shared hit=24
--   ->  Seq Scan on public.orders  (cost=0.00..23583.00 rows=248233 width=58) (actual time=0.010..0.370 rows=499 loops=1)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.status)::text = 'returned'::text)
--         Rows Removed by Filter: 1624
--         Buffers: shared hit=24
-- Planning Time: 0.068 ms
-- Execution Time: 0.406 ms

-- HASH (cid, ppid) + BTREE (status)
-- Limit  (cost=74.96..80.47 rows=58 width=58) (actual time=1.080..1.169 rows=58 loops=1)
-- "  Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--   Buffers: shared hit=39
--   ->  Seq Scan on public.orders  (cost=0.00..23583.00 rows=248233 width=58) (actual time=0.010..1.129 rows=847 loops=1)
-- "        Output: id, client_id, pickup_point_id, status, status_updated, weight, cost"
--         Filter: ((orders.status)::text = 'returned'::text)
--         Rows Removed by Filter: 2585
--         Buffers: shared hit=39
-- Planning Time: 0.071 ms
-- Execution Time: 1.188 ms


-----------------------------------------------
-- В итоге самым оптимальным оказались hash индексы для client_id, pickup_point_id и status
-- (хотя добавление индекса на status не дало ощутимых улучшений)

