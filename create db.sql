DROP TABLE soldTv;
DROP TABLE tv;

CREATE TABLE tv(
    id integer PRIMARY KEY NOT NULL,
    brand string NULL,
    manufacturer string NOT NULL CHECK(length(manufacturer) >= 3),
    model string NOT NULL CHECK(length(model) >= 2),
    year integer CHECK(year >= 2010)
);

PRAGMA foreign_keys = ON; 

CREATE TABLE soldTv(
id Sqlite3_int64 PRIMARY KEY,
sold_count Sqlite3_int64 NOT NULL CHECK(sold_count >= 0),
available Sqlite3_int64 NOT NULL CHECK(available >= 0),
tv_id Sqlite3_int64 NOT NULL UNIQUE,
FOREIGN KEY(tv_id) REFERENCES tv(id)
);

INSERT INTO tv(brand, manufacturer, model, year)
VALUES
('samsung', 'korea', 'a1', 2011),
('lg', 'korea', 'qwe123', 2013),
(NULL,  'china', 'a21', 2015),
('samsung', 'korea', 'a2', 2012),
('samsung', 'korea', 'a3', 2011),
('samsung', 'korea', 'a4', 2011);

INSERT INTO soldTv(sold_count, available, tv_id)
VALUES
(5000, 14345, 1),
(123534512, 32543, 3),
(55, 100000, 6);

