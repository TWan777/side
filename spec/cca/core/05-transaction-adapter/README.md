# Transaction Adapter

To facilitate the connection of CCA applications with multiple blockchains, CCA applications need to implement a `Transaction Adapter`. This adapter serves to convert transactions into a format compatible with remote blockchains.

## Definitions

## Technical Specification

### Data Structures
```ts
interface EthereumTx {
    from: string,
    to: string,
    gas: string, // 30400,
    gasPrice: string, // 10000000000000
    value: string, // 2441406250
    data: string"
}
```

```ts
interface TxAdapter {
    toRemoteTx(chainType: ClientType, request: SignningRequest);
    verifyInboundTransaction(request: IntentRequest, tx: byte[]): bool;
}
```

