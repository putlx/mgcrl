CREATE TABLE assets (
	name TEXT PRIMARY KEY,
	url TEXT,
	version TEXT,
	last_chapter TEXT
);

CREATE TABLE config (
	attr TEXT PRIMARY KEY,
	val TEXT
);

INSERT INTO config VALUES ('max_retry', 3), ('duration', 5 * 60), ('output', '.'), ('freq_in_hour', 6);

CREATE TABLE ac_log (
	time INTEGER,
	message TEXT
);

CREATE TABLE sv_log (
	time INTEGER,
	message TEXT
);
