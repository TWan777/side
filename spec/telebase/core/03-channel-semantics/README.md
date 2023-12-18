# Channel

A channel functions as a conduit for interchain operations between a designated module on the SIDE blockchain and a CCA on a remote blockchain. It guarantees that the vault exclusively accepts transactions sent from the specified module. Additionally, it triggers an acknowledgment function of the module to update states by monitoring received transactions that interact with the CCA on the remote blockchain.

## Data Structure
```ts
interface Channel {
   // specific which client is used for verification of inbound transaction. 
   client_id: string;
   vault_address: string;
   app_id: string;
   // verify if this transaction is included in the external chain.
   function verify(txHash: string, proof: byte[]);
}
```


To prevent replay attacks, the states of both `IntentRequest` and `SigningRequest` transactions must be stored on the state chain.

```ts
interface IntentRequest {
   channelId: string
   action: string,
   expectedSender: string, // the expected sender of inboundTx on counterparty chain
   hash: string,
   status: Enum,
   inboundTx: bytes[],
   createAt: u64,
}

interface SigningRequest {
   channelId: string,
   action: string,  // can be defined in app, such as AtomicSwap, LSD
   hash: string,
   status: Enum,
   outboundTx: bytes[],
   createAt: u64,
}
```

To enhance security, `IntentRequests` have a time limit of 24 hours. Any assets deposited after this designated time frame will not be processed and, consequently, will be forfeited. Users must adhere to the 24-hour limit to ensure the successful completion of the asset deposit on the external chain. Additionally, only one open request is allowed per account.

There should be the following functions to facilitate obtaining transactions or iterating through transactions.

```ts
function getRequest(hash: string) {

}

function getRequests(status: Enum, start: int, limit: int) {

}
```

## Definition

 - **Inbound Transaction** : A transaction initiated by users that involves interaction with the Cross Chain Account (CCA) on a remote blockchain.
 - **Outbound Transaction**: A transaction initiated based on users' IntentRequest on the controller chain. This request involves the TSS network for signing and execution on the remote chain.

## Life Scope

**Life Scope of Inbound Transaction:**
1. **Registered:** The IntentRequest is registered on the controller chain.
2. **Executed:** The transaction has been executed on the remote chain, and the result is synced to the controller chain by a relayer.
3. **Confirmed:** The transaction has passed validation by the light client on the controller chain.
4. **Finalized:** The transaction has been finalized on the controller chain.
5. **Expired:** An IntentRequest that hasn't received the corresponding inbound transaction within a fixed window.

**Life Scope of Outbound Transaction:**
1. **Initiated:** A SigningRequest is created by a transaction on the controller chain.
2. **Signed:** The transaction has been signed by the TSS network.
3. **Broadcasted:** The transaction has been written to a relay queue for broadcasting.
4. **Executed:** The transaction has been executed on the remote chain, and the result is synced to the controller chain by a relayer.
5. **Confirmed:** The transaction has passed validation by the light client on the controller chain.
6. **Finalized:** The transaction has been finalized on the controller chain.

