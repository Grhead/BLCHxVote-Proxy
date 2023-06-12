-- +goose Up
-- +goose StatementBegin
CREATE TABLE AuthDataTable (
                            Id          CHAR(36) PRIMARY KEY,
                            Login       CHAR(64) NOT NULL,
                            Password    CHAR(64) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query';
