# Relayer

To simplify the implementation, there should be at most one bridge connecting two blockchains. This implies that only one vault should exist on the counterparty chain.

The `pendingTransactions` function computes the set of all valid transactions that need to be relayed from one chain to another, taking into account the state of both chains.

The `submitTransaction` function is chain-specific, involving the submission of a transaction. Transactions can be submitted either individually as single transactions or atomically as a single transaction, depending on the capabilities of the chain.

The `relay` function is invoked at regular intervals but no more frequently than once per block on either chain.

```ts
function relay(c: Chain>) {
    const txs = chain.pendingTransactions(c.counterparty)
    for (const localTx of txs)
      chain.submitDatagram(localTx)
}
```
