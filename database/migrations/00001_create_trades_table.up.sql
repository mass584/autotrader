create table trades (
	id bigint unsigned primary key auto_increment,
	exchange_place tinyint unsigned not null,
	exchange_pair tinyint unsigned not null,
	trade_id bigint unsigned not null,
	price decimal(20, 10) not null,
	volume decimal(20, 10) not null,
    time datetime not null,
	created_at timestamp not null default CURRENT_TIMESTAMP,
	updated_at timestamp not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,
	unique index idx_exchange_place_exchange_pair_trade_id (exchange_place, exchange_pair, trade_id),
	index idx_exchange_place_exchange_pair_time (exchange_place, exchange_pair, time)
)