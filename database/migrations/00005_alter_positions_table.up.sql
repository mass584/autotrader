alter table positions
modify column buy_price decimal(20, 10) null,
modify column sell_price decimal(20, 10) null,
modify column buy_time datetime null,
modify column sell_time datetime null;