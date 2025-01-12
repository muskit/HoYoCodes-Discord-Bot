CREATE DATABASE IF NOT EXISTS scraper;
USE scraper;

CREATE TABLE `Codes` (
  `code` varchar(50),
  `game` ENUM ('Honkai Impact 3rd', 'Genshin Impact', 'Honkai Star Rail', 'Zenless Zone Zero'),
  `description` text,
  `found` datetime,
  `is_livestream` bool,
  PRIMARY KEY (`code`, `game`)
);

CREATE TABLE `ScrapeStats` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `last_scrape_successful` bool,
  `last_scrape_when` datetime,
  `last_edit_when` datetime
);

