create table hole
(
    id              int not null auto_increment,
    course_id       int not null,
    number          int not null,
    par             int not null,
    stroke          int not null,
    distance_yards  int not null,
    distance_meters int not null,
    primary key (id),
    constraint hole_course_id_fk
        foreign key (course_id) references course (id)
);

