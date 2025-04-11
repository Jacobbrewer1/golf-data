create table course_details
(
    id                int           not null auto_increment,
    course_id         int           not null,
    marker            varchar(255)  not null,
    slope_rating      int           not null,
    course_rating     decimal(4, 1) not null,
    par_front_nine    int           not null,
    par_back_nine     int           not null,
    par_total         int           not null,
    yards_front_nine  int           not null,
    yards_back_nine   int           not null,
    yards_total       int           not null,
    meters_front_nine int           not null,
    meters_back_nine  int           not null,
    meters_total      int           not null,
    primary key (id),
    constraint course_details_course_id_fk
        foreign key (course_id) references course (id)
);

