alter table positions
modify column buy_price decimal(20, 10) not null,
modify column sell_price decimal(20, 10) not null,
modify column buy_time datetime not null,
modify column sell_time datetime not null;