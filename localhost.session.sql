CREATE TABLE ethereum_transactions
(
    tx_hash FixedString(66),
    block_number UInt64,
    block_hash FixedString(66),
    from_address FixedString(42),
    to_address Nullable(FixedString(42)),
    value Decimal(78, 0),
    input String,
    gas_price UInt64,
    gas_limit UInt64,
    gas_used Nullable(UInt64),
    nonce UInt64,
    v UInt64,
    r FixedString(66),
    s FixedString(66),
    timestamp DateTime,
    transaction_status UInt8,
    transaction_index UInt32

) ENGINE = ReplacingMergeTree()
ORDER BY (block_number, tx_hash)
SETTINGS index_granularity = 8192;
