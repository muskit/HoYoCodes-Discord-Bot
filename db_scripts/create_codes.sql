CREATE DATABASE Codes;
USE Codes;

-- Game no:
-- 0: Honkai Impact
-- 1: Genshin Impact
-- 2: Honkai Star Rail
-- 3: Zenless Zone Zero

CREATE TABLE Codes (
    id VARCHAR(100) PRIMARY KEY,
    game INT NOT NULL,
    is_active BOOLEAN NOT NULL,
    is_recent BOOLEAN NOT NULL -- based on datetime_updated
);

CREATE TABLE Update_Datetimes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    datetime_scraped DATETIME NOT NULL,
    datetime_updated DATETIME NOT NULL -- only update if new data's update datetime is within 4 hrs of stored!
);
