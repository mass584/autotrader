create table scraping_histories (
	id bigint unsigned primary key auto_increment,
	scraping_status tinyint unsigned not null,
	exchange_place tinyint unsigned not null,
	exchange_pair tinyint unsigned not null,
	from_id bigint unsigned not null,
	to_id bigint unsigned not null,
	from_time datetime not null,
	to_time datetime not null,
	created_at timestamp not null default CURRENT_TIMESTAMP,
	updated_at timestamp not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,
	index idx_exchange_place_exchange_pair (exchange_place, exchange_pair)
)