CREATE DATABASE secretary;

CREATE TABLE secretary.lab_events
(
    id         INTEGER                            NOT NULL PRIMARY KEY,
    username   VARCHAR(128)                       NOT NULL,
    event_type VARCHAR(128)                       NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);