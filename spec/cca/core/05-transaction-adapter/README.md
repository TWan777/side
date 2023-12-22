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
    data: string,
}

interface EthereumTxResponse {
    hash : string,
    nonce: u32,
    block_hash : string,
    block_number:u64,
    transaction_index UInt16,
    from_address : string,
    to_address : string,
    value:u64,
    gas:u64,
    gas_price:u64,
    input : string,
    block_timestamp:u64,
    max_fee_per_gas:u64,
    max_priority_fee_per_gas:u64,
    transaction_type u8,
    receipt_cumulative_gas_used: u32,
    receipt_gas_used: u32,
    receipt_contract_address : string,
    receipt_root : string,
    receipt_status u8,
    receipt_effective_gas_price:u64
}
```

```ts
interface TxAdapter {
    buildSigningRequest(chainType: ClientType, request: SignningRequest);
    verifyInboundTransaction(request: IntentRequest, tx: byte[]): bool;
}

class DefaultEthereumAdapter {
    buildSigningRequest(action: string, channel: Channel, recipient: string, value: string, data: string): SignningRequest {
        const tx = {
            from: channel.getVaultAddress(),
            to: recipient,
            gas: channel.getDefaultGas(),
            gasPrice: channel.getDefaultGasPrice(),
            value,
            data,
        }
        return {
            channelId: channel.id,
            action, // can be defined in app, such as AtomicSwap, LSD
            hash: hash(tx),
            status: "CREATED",
            outboundTx: hex(tx),
            createAt: block.timestamp,
        }

    }
}

const ADAPTOR_REGISTRY: Record<ChainType, TxAdapter> = {
    Ethereum: new DefaultEthereumAdapter();
}
```

