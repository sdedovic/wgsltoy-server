CREATE TYPE email_verification_type as ENUM ('pending', 'completed');

CREATE TABLE IF NOT EXISTS users (
  user_id             SERIAL                    PRIMARY KEY      ,

  email               text                      UNIQUE NOT NULL  ,
  email_verification email_verification_type           NOT NULL  ,

  username            text                      UNIQUE NOT NULL  ,
  password            text                             NOT NULL  ,

  created_at          timestamp with time zone  
);
