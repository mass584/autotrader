create table trade_aggregations (
	id bigint unsigned primary key auto_increment,
	exchange_place tinyint unsigned not null,
	exchange_pair tinyint unsigned not null,
    aggregate_date date not null,
	average_price decimal(20, 10) not null,
	total_count bigint unsigned not null,
	total_transaction decimal(25, 10) not null,
	created_at timestamp not null default CURRENT_TIMESTAMP,
	updated_at timestamp not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,
	unique index idx_exchange_place_exchange_pair_aggregate_date (exchange_place, exchange_pair, aggregate_date)
)