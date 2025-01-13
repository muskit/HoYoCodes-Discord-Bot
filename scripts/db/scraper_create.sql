CREATE DATABASE IF NOT EXISTS scraper;
USE scraper;

CREATE TABLE `Codes` (
  `code` varchar(50),
  `game` ENUM ('Honkai Impact 3rd', 'Genshin Impact', 'Honkai Star Rail', 'Zenless Zone Zero'),
  `description` text,
  `added` datetime,
  `is_livestream` bool,
  PRIMARY KEY (`code`, `game`)
);

CREATE TABLE `ScrapeStats` (
  `game` ENUM ('Honkai Impact 3rd', 'Genshin Impact', 'Honkai Star Rail', 'Zenless Zone Zero'),
  `updated` datetime,
  `checked` datetime,
  PRIMARY KEY (`game`)
);

