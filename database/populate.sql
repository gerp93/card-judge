USE CARD_JUDGE;

INSERT INTO USER (NAME, PASSWORD_HASH, IS_ADMIN)
VALUES ('Grant', '$2a$14$t7gWxR3Ak8uBkyPnw4TZz.WcN3nVlbDMEQgqHOuxEfWN3yCL3dgY.', 1);

INSERT INTO CARD_TYPE (ID, NAME)
VALUES ('a907026b-6fa6-11ef-b1ac-3bd680fc6f38', 'Judge'),
       ('a90703f8-6fa6-11ef-b1ac-3bd680fc6f38', 'Player');
