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

DepositRequest

```ts
interface DepositRequest {

}
```

WithdrawRequest

```ts
interface WithdrawRequest {

}
```
