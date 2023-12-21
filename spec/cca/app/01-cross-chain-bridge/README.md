# Cross chain Bridge

The Cross-Chain Bridge is a CCA app that enables users to deposit assets from an external blockchain to mint Peggy assets on the contoller blockchain. It also facilitates the burning of Peggy assets on the controller blockchain to withdraw native assets on the external blockchain.

Similar to many other bridge solutions, we wrap bridged assets into pegged assets with a 1:1 ratio. Users have the flexibility to mint pegged assets by initiating a `IntentRequest` or burn pegged assets through the execution of a `SigningRequest` to withdraw native assets.

## Definition

 - `Peggy Token`: A tokenized asset pegged on the controller blockchain, enabling smooth interoperability. This involves locking assets on a counterparty remote chain and minting equivalent tokens on the contoller blockchain, facilitating seamless cross-chain asset transfers while maintaining a fixed value ratio. For consistency and clarity, all peggy tokens must adhere to the naming convention: they should commence with the prefix `side/` followed by a hash generated from `channel_id`, `CCA address`, and `symbol`. This ensures a standardized and identifiable nomenclature for peggy tokens.
 - A `DepositRequest` is a specialized form of `IntentRequest` that triggers a request enabling the remote CCA account to receive and lock a specified amount of tokens. These tokens can then be utilized to mint a voucher token on the controller chain.
 - A `WithdrawRequest` is a specialized type of `SigningRequest` that involves burning voucher tokens. It initiates a request to the TSS network to sign a transfer transaction on the remote chain, allowing users to withdraw their locked native tokens from the remote chain.

## Technical Specification
### Transaction Flow 
![flow](./bridge_workflow.png)

### Data Structure

Transaction Registry
```ts
const Record<string, Record<ChainType, object>> txRegistry = {
    "SEND" {
        "Ethereum": {
          "from": "0xb60e8dd61c5d32be8058bb8eb970870f07233155",
          "to": "0xd46e8dd67c5d32be8058bb8eb970870f07244567",
          "gas": "0x76c0", // 30400,
          "gasPrice": "0x9184e72a000", // 10000000000000
          "value": "0x9184e72a", // 2441406250
          "data": "0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675"
        },
        "Bitcoin": {
            // ...
        },
    }
};
```

DepositRequest
```ts
interface DepositRequest extends IntentRequest{

}
```

WithdrawRequest

```ts
interface WithdrawTxAdapter {
    key: "Send",
    toRemoteTx(chain: ChainType): byte[] {
        const tx = txRegistry[this.key][chain];
        // assign request value to tx. 
    },
}
interface WithdrawRequest extends SigningRequest{
    adapter: WithdrawTxAdapter;
}
```
