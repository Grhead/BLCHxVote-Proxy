-- +goose Up
-- +goose StatementBegin
CREATE TABLE MasterUserRelations (
                                     Id        CHAR(36) PRIMARY KEY,
                                     UserId    CHAR(36) NOT NULL,
                                     MasterId CHAR(36) NOT NULL,
                                     FOREIGN KEY(MasterId) REFERENCES ChainMasters(Id),
                                     FOREIGN KEY(UserId) REFERENCES AuthDataTable(Id)
);
CREATE TABLE ChainMasters (
                              Id            CHAR(36) PRIMARY KEY,
                              MasterChain   CHAR(64) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
SELECT 'down SQL query';

