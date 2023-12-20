# Transaction Adapter

To facilitate the connection of CCA applications with multiple blockchains, CCA applications need to implement a `Transaction Adapter`. This adapter serves to convert transactions into a format compatible with remote blockchains.

## Definitions

## Technical Specification

### Data Structures

```ts
interface TxAdapter {
    toRemoteTx(chainId: string, request: SignningRequest);
}
```

