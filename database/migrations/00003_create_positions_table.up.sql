create table positions (
	id bigint unsigned primary key auto_increment,
	position_status tinyint unsigned not null,
	exchange_place tinyint unsigned not null,
	exchange_pair tinyint unsigned not null,
	price decimal(20, 10) not null,
	volume decimal(20, 10) not null,
	time datetime not null,
	created_at timestamp not null default CURRENT_TIMESTAMP,
	updated_at timestamp not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,
	index idx_exchange_place_exchange_pair (exchange_place, exchange_pair)
)