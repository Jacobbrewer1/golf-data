create table course_details
(
    id                int           not null auto_increment,
    course_id         int           not null,
    marker            varchar(255) null,
    slope             int           not null,
    course_rating     decimal(4, 1) not null,
    front_nine_par    int           not null,
    back_nine_par     int           not null,
    total_par         int           not null,
    front_nine_yards  int           not null,
    back_nine_yards   int           not null,
    total_yards       int           not null,
    front_nine_meters int           not null,
    back_nine_meters  int           not null,
    total_meters      int           not null,
    primary key (id),
    constraint course_details_course_id_fk
        foreign key (course_id) references course (id)
);

