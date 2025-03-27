create table club
(
    id          int          not null auto_increment,
    name        varchar(255) not null,
    address1    varchar(255) not null,
    address2    varchar(255) not null,
    address3    varchar(255) not null,
    address4    varchar(255) not null,
    postal_code varchar(8)   not null,
    primary key (id),
    unique (name)
);

