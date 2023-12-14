### Light Client

The light client traces states on counterparty chains and can be implemented in two types: probabilistic finality for PoW consensus blockchains (e.g., Bitcoin) and deterministic finality for PoS blockchains (e.g., Cosmos and Ethereum). 

The light client on the SIDE blockchain is responsible for verifying that a inbound transaction has been executed on the counterparty blockchain. Its primary role is to ensure that the vault account has received the necessary assets as a result of the transaction.

#### Data Structure

 - `Client State`
```ts
interface ClientState {
   chainId: string,
   type: string,
   latestHeader: Header,
   comfirmation: u64,
   frozen: bool,
   validators: Vec<pubkey, u64>, // only used for PoS client
   difficulty: u64, // only used for PoW client
}
```
 - `Header`
```ts
interface Header {
    height: u64,
    hash: string,
    previous_hash: string,
    root: string,
}
```
 - Initilize Client

The Light Client can be initialized through on-chain governance by specifying parameters such as `clientId`, a trusted `Header`, `vault address`, and other relevant details.

```ts
function initialClient(identifier: string, clientState: ClientState, header: Header) {}
```
 
 - Update Client

The relayer periodically updates the latest state to the on-chain light client. The update frequency depends on the connected blockchain, and failure to update in a timely manner may result in the light client being unable to process the latest transactions.

The client state is fundamental to the security assumptions, and a vulnerable implementation poses risks to the bridge, increasing the potential for losing assets.

For **PoS consensus** light clients, verification includes the following conditions:
   1. The block should have a minimum threshold of voting power signatures from trusted validators.
   2. `chain_id` should be the same.
   3. `height` should be greater than the current height.
   4. The client state can be updated along with the inbound transaction or if the validator set has changed.

For **PoW consensus** light clients, verification includes the following conditions:
   1. Check if the block hash matches the difficulty.
   2. Check if the block includes the hash of the latest trusted block (height-1).
   3. `chain_id` should be the same.
   4. `height` should be greater than the current height.
   5. The client state should be updated at each height.
   6. Headers should be allowed to override before confirmation since the longest blockchain might have a different height than the shorter one.


 ```ts
function updateClient(identifier: string, clientState: ClientState, header: Header) {}
```

 - Verify Transaction

Transactions can be verified by checking the following conditions using the on-chain light client:

   1. Verifying if the transaction is included in the Merkle tree root of the block header.
   2. Verifying if the chain ID is the same.
   3. Verifying if the receiver is the vault address.
   4. Verifying if the transaction has executed successfully.
   5. Verifying if the transaction has deposited sufficient assets.
   6. Verifying if the transaction was executed within a designated time frame from the request start.

The relayer is responsible for generating the proof of inclusion, which helps ensure the integrity and validity of the transactions.

```ts
function verifyTransaction(identifer: string, header: Header, txHash: string, proof: byte[]) {}
```

#### Implementations

 - Bitcoin Light Client
 - Ethereum Light Client
 - BSC Light Client
 - Solana Light Client
