 -- CREATE TABLE "news" -----------------------------------------
CREATE TABLE `news` ( 
	`id` BigInt( 255 ) AUTO_INCREMENT NOT NULL,
	`author` Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
	`body` Text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
	`created` Timestamp NOT NULL,
	CONSTRAINT `unique_id` UNIQUE( `id` ) )
CHARACTER SET = utf8mb4
COLLATE = utf8mb4_unicode_ci
ENGINE = InnoDB;
-- -------------------------------------------------------------

