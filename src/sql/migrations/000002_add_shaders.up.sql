CREATE TYPE shader_visibility_type as ENUM  ('private', 'unlisted', 'public');

CREATE TABLE IF NOT EXISTS shaders (
    shader_id           character(22)                                 PRIMARY KEY  ,
    created_at          timestamp with time zone                      NOT NULL     ,
    updated_at          timestamp with time zone                      NOT NULL     ,

    created_by          character(22) REFERENCES users (user_id)      NOT NULL     ,
    visibility          shader_visibility_type                        NOT NULL     ,
    name                text                                          NOT NULL     ,
    description         text                                          NOT NULL     ,
    content             text                                          NOT NULL     ,
    tags                text[]                                        NOT NULL
);