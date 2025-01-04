SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS GuildAdmins;
CREATE TABLE `GuildAdmins` (
  `guild_id` bigint unsigned,
  `role_id` bigint unsigned,
  PRIMARY KEY (`guild_id`, `role_id`)
);

DROP TABLE IF EXISTS Subscriptions;
CREATE TABLE `Subscriptions` (
  `id` bigint unsigned PRIMARY KEY AUTO_INCREMENT,
  `channel_id` bigint unsigned,
  `game` ENUM ('Honkai_Impact', 'Genshin_Impact', 'Honkai_Star_Rail', 'Zenless_Zone_Zero') COMMENT 'If null, notify for all games.',
  `ping_on_code_add` bool DEFAULT true,
  `ping_on_code_remove` bool DEFAULT false
);

DROP TABLE IF EXISTS SubscriptionRolePings;
CREATE TABLE `SubscriptionRolePings` (
  `subscription_id` bigint unsigned,
  `role_id` bigint unsigned,
  PRIMARY KEY (`subscription_id`, `role_id`)
);

DROP TABLE IF EXISTS Embeds;
CREATE TABLE `Embeds` (
  `message_id` bigint unsigned,
  `game` ENUM ('Honkai_Impact', 'Genshin_Impact', 'Honkai_Star_Rail', 'Zenless_Zone_Zero') COMMENT 'If null, notify for all games and ensure message_id is unique in table.',
  PRIMARY KEY (`message_id`, `game`)
);

ALTER TABLE `SubscriptionRolePings` ADD FOREIGN KEY (`subscription_id`) REFERENCES `Subscriptions` (`id`);
SET FOREIGN_KEY_CHECKS = 1;