package repository

import "github.com/Masterminds/squirrel"

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
