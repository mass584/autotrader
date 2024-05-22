create table trades (
	id bigint primary key auto_increment,
	exchange_name varchar(255) not null,
	trade_id varchar(255) not null,
	price decimal(20, 10) not null,
	volume decimal(20, 10) not null,
    time datetime not null,
	created_at timestamp default CURRENT_TIMESTAMP,
	updated_at timestamp default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP,
	unique (exchange_name, trade_id)
)