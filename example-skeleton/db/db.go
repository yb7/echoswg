package db

import (
    "context"
    "database/sql"

    "github.com/yb7/alilog"

    "github.com/yb7/echoswg/example-skeleton/config"
    _ "github.com/jackc/pgx/v5/stdlib"
)

var SQLDB *sql.DB

func OpenDB() {
    db, err := sql.Open("pgx", config.C.Db.Url)
    if err != nil {
        alilog.Fatal(err)
    }
    SQLDB = db
}

func DbMigrate() {
    if SQLDB == nil {
        alilog.Warnf("skip db migrate because SQLDB is nil")
    }
}

func CloseDB() {
    if SQLDB != nil {
        _ = SQLDB.Close()
    }
}

func WithTx(ctx context.Context, fn func(ctxInTx context.Context, tx *sql.Tx) error) error {
    if SQLDB == nil {
        return alilog.Errorf("sql db is nil")
    }
    tx, err := SQLDB.BeginTx(ctx, nil)
    if err != nil {
        return alilog.Errorf("start transaction err: %v", err)
    }
    defer func() {
        if v := recover(); v != nil {
            _ = tx.Rollback()
            panic(v)
        }
    }()
    if err := fn(ctx, tx); err != nil {
        if rerr := tx.Rollback(); rerr != nil {
            return alilog.Errorf("rollback transaction err: %v", rerr)
        }
        return err
    }
    if err := tx.Commit(); err != nil {
        return alilog.Errorf("commit transaction err: %v", err)
    }
    return nil
}
