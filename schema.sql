-- This stores info about users -- Rupert
CREATE TABLE users (
    id INT PRIMARY KEY
);

-- This table stores high level information about the tours themselves
CREATE TABLE tours (
   id INTEGER PRIMARY KEY, -- this is the unique id of the tour that is given by sqlite automatically
   tour_name CHARACTER(100) NOT NULL, -- this is the name of the tour that the user specifies
   description TEXT, -- human readable description of the tour, can be nullable
   creator_user_number INT, -- id of the user who created this tour
   create_tsp TEXT -- date/time of creation
);

/*
This table contains the list of segments (and their order) for each tour
there is one or more segments for the tour_id
*/


CREATE TABLE tour_segments (
    tour_id INT REFERENCES tours(id) ON UPDATE CASCADE, -- if the tour id is changed, take me with you.
    segment_id INT, -- strava segment_id,
    PRIMARY KEY (tour_id, segment_id) -- you can't have the same segment in a single tour twice.
);


CREATE TABLE tour_members (
       tour_id INT REFERENCES tours(id) on UPDATE CASCADE, -- the tour this person is a rider on
       athlete_id INT,
       PRIMARY KEY (tour_id, athlete_id)

);
-- Now create a tour with some dummy data

insert into tours (id, tour_name, description, creator_user_number, create_tsp) VALUES
       ( 1, "Tour of NY Bridges", "Traverse the Williamsburg, Brooklyn, Manhattan and Ed Koch bridges in the fastest cumulative time.", 3004839, "2016-08-23T00:21:19+00:00");

insert into tour_segments (tour_id, segment_id) VALUES (1, 3377798); -- williamsburg
insert into tour_segments (tour_id, segment_id) VALUES (1, 2622770); -- manhattan
insert into tour_segments (tour_id, segment_id) VALUES (1, 640095); -- brooklyn
insert into tour_segments (tour_id, segment_id) VALUES (1, 2276683); -- ed koch
