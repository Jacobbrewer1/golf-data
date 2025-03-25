create table course
(
    id      int          not null auto_increment,
    club_id int          not null,
    name    varchar(255) not null,
    primary key (id),
    constraint course_club_id_fk
        foreign key (club_id) references club (id)
);

