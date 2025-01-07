DROP TABLE IF EXISTS SubscriptionGames;
DROP TABLE IF EXISTS SubscriptionPingRoles;
DROP TABLE IF EXISTS Subscriptions;
DROP TABLE IF EXISTS Embeds;

CREATE TABLE `Subscriptions` (
  `channel_id` BIGINT UNSIGNED PRIMARY KEY,
  `ping_on_code_add` bool DEFAULT true,
  `ping_on_code_remove` bool DEFAULT false
);

CREATE TABLE `SubscriptionGames` (
  `channel_id` BIGINT UNSIGNED PRIMARY KEY,
  `game` ENUM ('Honkai Impact', 'Genshin Impact', 'Honkai Star Rail', 'Zenless Zone Zero')
);

CREATE TABLE `SubscriptionPingRoles` (
  `channel_id` BIGINT UNSIGNED,
  `role_id` BIGINT UNSIGNED,
  PRIMARY KEY (`channel_id`, `role_id`)
);

CREATE TABLE `Embeds` (
  `message_id` BIGINT UNSIGNED,
  `game` ENUM ('Honkai Impact', 'Genshin Impact', 'Honkai Star Rail', 'Zenless Zone Zero') COMMENT 'If null, notify for all games and ensure message_id is unique in table.',
  PRIMARY KEY (`message_id`, `game`)
);

ALTER TABLE `Subscriptions` ADD FOREIGN KEY (`channel_id`) REFERENCES `SubscriptionGames` (`channel_id`);

ALTER TABLE `Subscriptions` ADD FOREIGN KEY (`channel_id`) REFERENCES `SubscriptionPingRoles` (`channel_id`);

