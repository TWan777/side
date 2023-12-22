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
```

```ts
interface TxAdapter {
    buildSigningRequest(chainType: ClientType, request: SignningRequest);
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
    },
}

const TX_REGISTRY: Record<ChainType, TxAdapter> = {
    Ethereum: new DefaultEthereumAdapter();
}
```

```ts

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

interface TxResponseAdapter<TxResponse> {
    construct(chainType: ClientType, request: IntentRequest);
    txResponse: TxResponse;
    verify(): bool;
}

class DefaultEthereumResponseAdapter<State> extends TxResponseAdapter<EthereumTxResponse>  {
    txResponse: TxResponse;
    request: IntentRequest;
    state: State; // used for verification in sub class
    construct(request: IntentRequest, channel: Channel, state: State) {
        this.txResponse = JSON.parse(request.inboundTx);
        this.request = request;
        this.state = state
    };
    /// This verify function only check if CCA received a transaction.
    /// More verifictions should be checked on apps
    verify(): bool {
        // The light client has checked if this tx was successfully executed on the remote chain
        if(this.txResponse.from_address !== request.expectedSender) return false
        if(this.txResponse.to_address !== channel.vaultAddress) return false
        return true;
    };
}

```
