# Channel


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
