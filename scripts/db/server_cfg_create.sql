# for MariaDB/MySQL
USE server_cfg;

CREATE TABLE `Subscriptions` (
  `channel_id` BIGINT UNSIGNED PRIMARY KEY,
  `guild_id` BIGINT UNSIGNED COMMENT 'For server-wide config checking.',
  `active` BOOL DEFAULT true,
  `ping_on_code_add` BOOL DEFAULT true,
  `ping_on_code_remove` BOOL DEFAULT false
);

CREATE TABLE `SubscriptionGames` (
  `channel_id` BIGINT UNSIGNED,
  `game` ENUM ('Honkai Impact 3rd', 'Genshin Impact', 'Honkai Star Rail', 'Zenless Zone Zero'),
  PRIMARY KEY (`channel_id`, `game`)
);

CREATE TABLE `SubscriptionPingRoles` (
  `channel_id` BIGINT UNSIGNED,
  `role_id` BIGINT UNSIGNED,
  PRIMARY KEY (`channel_id`, `role_id`)
);

CREATE TABLE `Embeds` (
  `message_id` BIGINT UNSIGNED,
  `game` ENUM ('Honkai Impact 3rd', 'Genshin Impact', 'Honkai Star Rail', 'Zenless Zone Zero') COMMENT 'If null, notify for all games and ensure message_id is unique in table.',
  PRIMARY KEY (`message_id`, `game`)
);

CREATE INDEX `subscription_guild_index` ON `Subscriptions` (`guild_id`);

ALTER TABLE `SubscriptionGames` ADD FOREIGN KEY (`channel_id`) REFERENCES `Subscriptions` (`channel_id`);

ALTER TABLE `SubscriptionPingRoles` ADD FOREIGN KEY (`channel_id`) REFERENCES `Subscriptions` (`channel_id`);
