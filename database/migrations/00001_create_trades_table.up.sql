create table trades (
	id varchar(255) primary key,
	price decimal(20, 10) not null,
	volume decimal(20, 10) not null,
    time datetime not null,
	created_at timestamp default CURRENT_TIMESTAMP,
	updated_at timestamp default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP
)